package main

import (
	"archive/zip"
	"os"
	"path/filepath"
	"testing"
)

func TestAddLibs(t *testing.T) {
	dir := t.TempDir()
	apk := filepath.Join(dir, "app.apk")
	f, _ := os.Create(apk)
	w := zip.NewWriter(f)
	for _, name := range []string{"AndroidManifest.xml", "lib/arm64-v8a/libgio.so", "META-INF/CERT.SF"} {
		e, _ := w.Create(name)
		e.Write([]byte(name))
	}
	w.Close()
	f.Close()
	libs := filepath.Join(dir, "libs")
	for _, p := range []string{"arm64-v8a/libsherpa-onnx-c-api.so", "arm64-v8a/libonnxruntime.so", "x86_64/libonnxruntime.so"} {
		os.MkdirAll(filepath.Dir(filepath.Join(libs, p)), 0o755)
		os.WriteFile(filepath.Join(libs, p), []byte(p), 0o644)
	}
	out := filepath.Join(dir, "out.apk")
	n, err := addLibs(apk, out, libs)
	if err != nil {
		t.Fatal(err)
	}
	if n != 2 {
		t.Fatalf("added %d, want 2 (only the APK's ABI)", n)
	}
	r, err := zip.OpenReader(out)
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()
	got := map[string]uint16{}
	for _, e := range r.File {
		got[e.Name] = e.Method
	}
	if _, ok := got["META-INF/CERT.SF"]; ok {
		t.Error("old signature kept")
	}
	if m, ok := got["lib/arm64-v8a/libonnxruntime.so"]; !ok || m != zip.Store {
		t.Errorf("library missing or compressed: %v", got)
	}
	if _, ok := got["lib/x86_64/libonnxruntime.so"]; ok {
		t.Error("added an ABI the APK wasn't built for")
	}
}
