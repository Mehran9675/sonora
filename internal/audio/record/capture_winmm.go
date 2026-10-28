//go:build windows

package record

import (
	"fmt"
	"sync"
	"unsafe"

	"golang.org/x/sys/windows"
)

// winmm captures with the classic waveIn API, which every Windows
// version has and which needs no cgo, so ARM64 builds stay simple.
var (
	winmm                = windows.NewLazySystemDLL("winmm.dll")
	procWaveInGetNumDevs = winmm.NewProc("waveInGetNumDevs")
	procWaveInGetDevCaps = winmm.NewProc("waveInGetDevCapsW")
	procWaveInOpen       = winmm.NewProc("waveInOpen")
	procWaveInPrepare    = winmm.NewProc("waveInPrepareHeader")
	procWaveInUnprepare  = winmm.NewProc("waveInUnprepareHeader")
	procWaveInAddBuffer  = winmm.NewProc("waveInAddBuffer")
	procWaveInStart      = winmm.NewProc("waveInStart")
	procWaveInStop       = winmm.NewProc("waveInStop")
	procWaveInReset      = winmm.NewProc("waveInReset")
	procWaveInClose      = winmm.NewProc("waveInClose")
)

const (
	waveMapper    = 0xFFFFFFFF
	callbackEvent = 0x00050000
	whdrDone      = 0x00000001
	bufferCount   = 4
	bufferBytes   = SampleRate / 10 * 2 // 100 ms of 16-bit mono
)

type waveFormatEx struct {
	FormatTag      uint16
	Channels       uint16
	SamplesPerSec  uint32
	AvgBytesPerSec uint32
	BlockAlign     uint16
	BitsPerSample  uint16
	Size           uint16
}

// waveInCaps is WAVEINCAPSW.
type waveInCaps struct {
	Mid, Pid      uint16
	DriverVersion uint32
	Pname         [32]uint16
	Formats       uint32
	Channels      uint16
	Reserved      uint16
}

type waveHdr struct {
	Data          uintptr
	BufferLength  uint32
	BytesRecorded uint32
	User          uintptr
	Flags         uint32
	Loops         uint32
	Next          uintptr
	Reserved      uintptr
}

type waveIn struct {
	handle  uintptr
	event   windows.Handle
	hdrs    [bufferCount]*waveHdr
	bufs    [bufferCount][]byte
	push    func([]byte)
	quit    chan struct{}
	done    sync.WaitGroup
	stopped bool
}

func call(p *windows.LazyProc, args ...uintptr) error {
	if r, _, _ := p.Call(args...); r != 0 {
		return fmt.Errorf("%s failed (MMRESULT %d)", p.Name, r)
	}
	return nil
}

// Devices lists the microphones by their (at most 31-character) names.
func Devices() ([]Device, error) {
	n, _, _ := procWaveInGetNumDevs.Call()
	var out []Device
	for i := uintptr(0); i < n; i++ {
		if name := deviceName(i); name != "" {
			out = append(out, Device{ID: name, Name: name})
		}
	}
	return out, nil
}

// deviceName is the name of input i, or "".
func deviceName(i uintptr) string {
	var caps waveInCaps
	if call(procWaveInGetDevCaps, i, uintptr(unsafe.Pointer(&caps)), unsafe.Sizeof(caps)) != nil {
		return ""
	}
	return windows.UTF16ToString(caps.Pname[:])
}

// deviceIndex returns the input called name, or the wave mapper (the
// default microphone) when there's none.
func deviceIndex(name string) uintptr {
	if name != "" {
		n, _, _ := procWaveInGetNumDevs.Call()
		for i := uintptr(0); i < n; i++ {
			if deviceName(i) == name {
				return i
			}
		}
	}
	return waveMapper
}

func open(device string, push func([]byte)) (capture, error) {
	ev, err := windows.CreateEvent(nil, 0, 0, nil)
	if err != nil {
		return nil, err
	}
	w := &waveIn{event: ev, push: push, quit: make(chan struct{})}
	format := waveFormatEx{FormatTag: 1, Channels: 1, SamplesPerSec: SampleRate,
		AvgBytesPerSec: SampleRate * 2, BlockAlign: 2, BitsPerSample: 16}
	if err := call(procWaveInOpen, uintptr(unsafe.Pointer(&w.handle)), deviceIndex(device),
		uintptr(unsafe.Pointer(&format)), uintptr(ev), 0, callbackEvent); err != nil {
		windows.CloseHandle(ev)
		return nil, fmt.Errorf("%w (%v)", ErrNoInput, err)
	}
	for i := range w.hdrs {
		w.bufs[i] = make([]byte, bufferBytes)
		w.hdrs[i] = &waveHdr{Data: uintptr(unsafe.Pointer(&w.bufs[i][0])), BufferLength: bufferBytes}
		if err := w.queue(w.hdrs[i], true); err != nil {
			w.stop()
			return nil, fmt.Errorf("%w (%v)", ErrNoInput, err)
		}
	}
	if err := call(procWaveInStart, w.handle); err != nil {
		w.stop()
		return nil, fmt.Errorf("%w (%v)", ErrNoInput, err)
	}
	w.done.Add(1)
	go w.loop()
	return w, nil
}

func (w *waveIn) queue(h *waveHdr, prepare bool) error {
	size := unsafe.Sizeof(*h)
	if prepare {
		if err := call(procWaveInPrepare, w.handle, uintptr(unsafe.Pointer(h)), size); err != nil {
			return err
		}
	}
	h.Flags &^= whdrDone
	return call(procWaveInAddBuffer, w.handle, uintptr(unsafe.Pointer(h)), size)
}

// loop hands filled buffers to push and queues them again.
func (w *waveIn) loop() {
	defer w.done.Done()
	for {
		windows.WaitForSingleObject(w.event, 200)
		select {
		case <-w.quit:
			return
		default:
		}
		for i, h := range w.hdrs {
			if h.Flags&whdrDone == 0 {
				continue
			}
			if n := h.BytesRecorded; n > 0 {
				buf := make([]byte, n)
				copy(buf, w.bufs[i][:n])
				w.push(buf)
			}
			h.BytesRecorded = 0
			w.queue(h, false)
		}
	}
}

func (w *waveIn) stop() {
	if w.stopped {
		return
	}
	w.stopped = true
	close(w.quit)
	windows.SetEvent(w.event)
	w.done.Wait()
	call(procWaveInStop, w.handle)
	call(procWaveInReset, w.handle)
	for _, h := range w.hdrs {
		if h != nil {
			call(procWaveInUnprepare, w.handle, uintptr(unsafe.Pointer(h)), unsafe.Sizeof(*h))
		}
	}
	call(procWaveInClose, w.handle)
	windows.CloseHandle(w.event)
}
