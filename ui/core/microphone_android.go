package core

import (
	"context"
	"errors"
	"time"

	"gioui.org/app"
	_ "gioui.org/app/permission/microphone" // RECORD_AUDIO in the manifest
	"git.wow.st/gmp/jni"
)

const recordAudio = "android.permission.RECORD_AUDIO"

// errMicDenied is returned when the user doesn't allow the microphone.
var errMicDenied = errors.New("Sonora needs the microphone to record. Allow it when Android asks, or in the app's settings")

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
