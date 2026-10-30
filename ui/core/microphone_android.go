package core

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"gioui.org/app"
	_ "gioui.org/app/permission/microphone" // RECORD_AUDIO in the manifest
	"git.wow.st/gmp/jni"

	"github.com/mehran9675/sonora/internal/audio/record"
	"github.com/mehran9675/sonora/ui/tr"
)

const recordAudio = "android.permission.RECORD_AUDIO"

// errMicDenied is returned when the user doesn't allow the microphone.
var errMicDenied = errors.New(tr.Msg("Sonora needs the microphone to record. Allow it when Android asks, or in the app's settings"))

// RequestMicrophone asks Android for the microphone permission, showing
// the system prompt the first time, and waits up to a minute for the
// answer. Gio doesn't hand the activity's permission result to apps, so
// the grant is polled. Call it off the UI goroutine.
func (c *Ctx) RequestMicrophone(ctx context.Context) error {
	vm := jni.JVMFor(app.JavaVM())
	if micGranted(vm) {
		return nil
	}
	view := c.App.NativeHandle()
	if view == 0 {
		return errMicDenied
	}
	var err error
	c.App.Window().Run(func() { err = requestMic(vm, view) })
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(ctx, time.Minute)
	defer cancel()
	t := time.NewTicker(300 * time.Millisecond)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return errMicDenied
		case <-t.C:
			if micGranted(vm) {
				return nil
			}
		}
	}
}

// micGranted reports whether the app holds RECORD_AUDIO.
func micGranted(vm jni.JVM) bool {
	granted := false
	jni.Do(vm, func(env jni.Env) error {
		appCtx := jni.Object(app.AppContext())
		check := jni.GetMethodID(env, jni.GetObjectClass(env, appCtx), "checkSelfPermission", "(Ljava/lang/String;)I")
		res, err := jni.CallIntMethod(env, appCtx, check, jni.Value(jni.JavaString(env, recordAudio)))
		granted = err == nil && res == 0 // PackageManager.PERMISSION_GRANTED
		return err
	})
	return granted
}

// requestMic shows the permission prompt through the activity that
// hosts view. Run it on the main thread.
func requestMic(vm jni.JVM, view uintptr) error {
	return jni.Do(vm, func(env jni.Env) error {
		v := jni.Object(view)
		getContext := jni.GetMethodID(env, jni.GetObjectClass(env, v), "getContext", "()Landroid/content/Context;")
		activity, err := jni.CallObjectMethod(env, v, getContext)
		if err != nil {
			return err
		}
		perms := jni.NewObjectArray(env, 1, jni.FindClass(env, "java/lang/String"), jni.Object(jni.JavaString(env, recordAudio)))
		request := jni.GetMethodID(env, jni.GetObjectClass(env, activity), "requestPermissions", "([Ljava/lang/String;I)V")
		return jni.CallVoidMethod(env, activity, request, jni.Value(perms), jni.Value(1))
	})
}

// inputTypes names the AudioDeviceInfo types that are microphones;
// others (telephony, tuners, echo reference…) are left out.
var inputTypes = map[int32]string{
	3:  tr.Msg("Wired headset"),     // TYPE_WIRED_HEADSET
	5:  tr.Msg("Line in"),           // TYPE_LINE_ANALOG
	6:  tr.Msg("Digital line in"),   // TYPE_LINE_DIGITAL
	7:  tr.Msg("Bluetooth headset"), // TYPE_BLUETOOTH_SCO
	11: tr.Msg("USB microphone"),    // TYPE_USB_DEVICE
	12: tr.Msg("USB accessory"),     // TYPE_USB_ACCESSORY
	13: tr.Msg("Dock"),              // TYPE_DOCK
	15: tr.Msg("Phone microphone"),  // TYPE_BUILTIN_MIC
	22: tr.Msg("USB headset"),       // TYPE_USB_HEADSET
	26: tr.Msg("Bluetooth headset"), // TYPE_BLE_HEADSET
}

// Microphones lists the inputs AudioManager reports; miniaudio sees
// only the default one on Android. It needs no permission.
func Microphones() ([]record.Device, error) {
	var out []record.Device
	seen := map[string]int{}
	err := jni.Do(jni.JVMFor(app.JavaVM()), func(env jni.Env) error {
		appCtx := jni.Object(app.AppContext())
		getService := jni.GetMethodID(env, jni.GetObjectClass(env, appCtx), "getSystemService", "(Ljava/lang/String;)Ljava/lang/Object;")
		am, err := jni.CallObjectMethod(env, appCtx, getService, jni.Value(jni.JavaString(env, "audio")))
		if err != nil {
			return err
		}
		getDevices := jni.GetMethodID(env, jni.GetObjectClass(env, am), "getDevices", "(I)[Landroid/media/AudioDeviceInfo;")
		arr, err := jni.CallObjectMethod(env, am, getDevices, jni.Value(1)) // GET_DEVICES_INPUTS
		if err != nil {
			return err
		}
		arrays := jni.FindClass(env, "java/lang/reflect/Array")
		getLength := jni.GetStaticMethodID(env, arrays, "getLength", "(Ljava/lang/Object;)I")
		n, err := jni.CallStaticIntMethod(env, arrays, getLength, jni.Value(arr))
		if err != nil {
			return err
		}
		for i := 0; i < n; i++ {
			info, err := jni.GetObjectArrayElement(env, jni.ObjectArray(arr), jni.Size(i))
			if err != nil {
				return err
			}
			cls := jni.GetObjectClass(env, info)
			id, err := jni.CallIntMethod(env, info, jni.GetMethodID(env, cls, "getId", "()I"))
			if err != nil {
				return err
			}
			kind, err := jni.CallIntMethod(env, info, jni.GetMethodID(env, cls, "getType", "()I"))
			if err != nil {
				return err
			}
			label, ok := inputTypes[kind]
			label = tr.T(label)
			if !ok {
				continue
			}
			// The product name tells headsets apart; built-in mics all
			// carry the phone's model, so they get a number instead.
			if kind != 15 {
				if name := javaText(env, info, cls, "getProductName"); name != "" {
					label += " · " + name
				}
			}
			seen[label]++
			if seen[label] > 1 {
				label = fmt.Sprintf("%s %d", label, seen[label])
			}
			out = append(out, record.Device{ID: record.AndroidPrefix + strconv.Itoa(int(id)), Name: label})
		}
		return nil
	})
	return out, err
}

// javaText calls a no-argument method returning a CharSequence and
// returns it as a string, or "".
func javaText(env jni.Env, obj jni.Object, cls jni.Class, method string) string {
	cs, err := jni.CallObjectMethod(env, obj, jni.GetMethodID(env, cls, method, "()Ljava/lang/CharSequence;"))
	if err != nil || cs == 0 {
		return ""
	}
	toString := jni.GetMethodID(env, jni.GetObjectClass(env, cs), "toString", "()Ljava/lang/String;")
	s, err := jni.CallObjectMethod(env, cs, toString)
	if err != nil || s == 0 {
		return ""
	}
	return jni.GoString(env, jni.String(s))
}
