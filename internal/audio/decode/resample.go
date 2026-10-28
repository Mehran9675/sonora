package decode

import "github.com/mehran9675/sonora/internal/audio"

// resampler converts a stream of mono blocks to its output rate as
// they arrive. Downsampling averages the input under each output
// sample (a box filter, enough to keep speech free of aliasing);
// upsampling interpolates linearly.
type resampler struct {
	rate int
	out  []float32

	in    int     // input rate, set by the first block
	ratio float64 // input samples per output sample
	n     int64   // input samples consumed
	next  float64 // input position of the next output sample
	sum   float64
	cnt   int
	last  float32
}

func newResampler(rate int) *resampler { return &resampler{rate: rate} }

func (r *resampler) empty() bool { return r.n == 0 }

// write takes a block of samples at rate in. It never fails; the error
// lets it serve as wav.Stream's callback.
func (r *resampler) write(in int, block []float32) error {
	if r.in == 0 {
		if r.rate == 0 { // keep the file's rate
			r.rate = in
		}
		r.in, r.ratio = in, float64(in)/float64(r.rate)
		if r.ratio > 1 {
			r.next = r.ratio
		}
	}
	for _, s := range block {
		switch {
		case r.ratio == 1:
			r.out = append(r.out, s)
		case r.ratio > 1:
			if float64(r.n) >= r.next {
				r.out = append(r.out, float32(r.sum/float64(r.cnt)))
				r.sum, r.cnt = 0, 0
				r.next += r.ratio
			}
			r.sum += float64(s)
			r.cnt++
		default:
			// Output samples between the previous input sample (at n-1)
			// and this one (at n).
			for r.next <= float64(r.n) {
				f := float32(r.next - float64(r.n-1))
				if r.n == 0 {
					f = 1
				}
				r.out = append(r.out, r.last+(s-r.last)*f)
				r.next += r.ratio
			}
		}
		r.last = s
		r.n++
	}
	return nil
}

func (r *resampler) clip() *audio.Clip {
	if r.cnt > 0 {
		r.out = append(r.out, float32(r.sum/float64(r.cnt)))
		r.cnt = 0
	}
	return &audio.Clip{Samples: r.out, SampleRate: r.rate}
}
