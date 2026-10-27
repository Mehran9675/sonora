// Command packapk adds the sherpa-onnx native libraries to an APK built
// by the giode CLI (which packs only the app's own libgio.so), then
// aligns and signs it again with the Android SDK's zipalign and
// apksigner.
//
//	go run ./tools/fetchsherpa android      # once: the libraries
//	giode build android                     # dist/android/Sonora.apk
//	go run ./tools/packapk dist/android/Sonora.apk
//
// Without -ks it signs with the SDK's debug key (~/.android/debug.keystore),
// for installing on your own phone. Releases pass their keystore, with
// its password in the environment variable named by -pass-env.
package main

import (
	"archive/zip"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

func main() {
	libs := flag.String("libs", filepath.Join("third_party", "sherpa-onnx", "android"), "folder of <abi>/*.so to add")
	out := flag.String("o", "", "output APK (default: replace the input)")
	ks := flag.String("ks", "", "keystore to sign with (default: the Android debug keystore)")
	passEnv := flag.String("pass-env", "", "environment variable holding the keystore password (default: the debug key's)")
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "usage: packapk [flags] app.apk")
		flag.PrintDefaults()
	}
	flag.Parse()
	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(2)
	}
	in := flag.Arg(0)
	if *out == "" {
		*out = in
	}
	tools, err := buildTools()
	if err != nil {
		log.Fatal(err)
	}
	unsigned := *out + ".unaligned"
	added, err := addLibs(in, unsigned, *libs)
	if err != nil {
		log.Fatal(err)
	}
	defer os.Remove(unsigned)
	log.Printf("added %d libraries", added)
	if err := run(filepath.Join(tools, "zipalign"), "-p", "-f", "4", unsigned, *out); err != nil {
		log.Fatal(err)
	}
	sign := []string{"sign"}
	if *ks == "" {
		home, _ := os.UserHomeDir()
		sign = append(sign, "--ks", filepath.Join(home, ".android", "debug.keystore"), "--ks-pass", "pass:android")
	} else {
		if *passEnv == "" {
			log.Fatal("-ks needs -pass-env")
		}
		sign = append(sign, "--ks", *ks, "--ks-pass", "env:"+*passEnv)
	}
	apksigner := filepath.Join(tools, "apksigner")
	if err := run(apksigner, append(sign, *out)...); err != nil {
		log.Fatal(err)
	}
	if err := run(apksigner, "verify", *out); err != nil {
		log.Fatal(err)
	}
	log.Printf("wrote %s", *out)
}

// addLibs copies the APK at in to out without its signature, adding
// every <abi>/*.so under libs for the ABIs the APK was built for. The
// libraries are stored uncompressed, as Android loads them in place.
func addLibs(in, out, libs string) (int, error) {
	r, err := zip.OpenReader(in)
	if err != nil {
		return 0, err
	}
	defer r.Close()
	f, err := os.Create(out)
	if err != nil {
		return 0, err
	}
	defer f.Close()
	w := zip.NewWriter(f)
	abis := map[string]bool{}
	have := map[string]bool{}
	for _, e := range r.File {
		if strings.HasPrefix(e.Name, "META-INF/") {
			continue // the old signature
		}
		if parts := strings.Split(e.Name, "/"); len(parts) == 3 && parts[0] == "lib" {
			abis[parts[1]] = true
		}
		have[e.Name] = true
		if err := w.Copy(e); err != nil {
			return 0, err
		}
	}
	if len(abis) == 0 {
		return 0, fmt.Errorf("%s has no native libraries; is it an APK from giode build android?", in)
	}
	var names []string
	for abi := range abis {
		found, _ := filepath.Glob(filepath.Join(libs, abi, "*.so"))
		if len(found) == 0 {
			return 0, fmt.Errorf("no libraries for %s in %s; run go run ./tools/fetchsherpa android", abi, libs)
		}
		names = append(names, found...)
	}
	sort.Strings(names)
	added := 0
	for _, src := range names {
		name := "lib/" + filepath.Base(filepath.Dir(src)) + "/" + filepath.Base(src)
		if have[name] {
			continue
		}
		if err := addStored(w, name, src); err != nil {
			return 0, err
		}
		added++
	}
	if err := w.Close(); err != nil {
		return 0, err
	}
	return added, f.Close()
}

func addStored(w *zip.Writer, name, src string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	dst, err := w.CreateHeader(&zip.FileHeader{Name: name, Method: zip.Store})
	if err != nil {
		return err
	}
	_, err = io.Copy(dst, in)
	return err
}

// buildTools returns the newest build-tools folder of the Android SDK.
func buildTools() (string, error) {
	sdk := os.Getenv("ANDROID_HOME")
	if sdk == "" {
		sdk = os.Getenv("ANDROID_SDK_ROOT")
	}
	if sdk == "" {
		return "", fmt.Errorf("set ANDROID_HOME to the Android SDK")
	}
	dirs, _ := filepath.Glob(filepath.Join(sdk, "build-tools", "*"))
	if len(dirs) == 0 {
		return "", fmt.Errorf("no build-tools in %s; install them with sdkmanager", sdk)
	}
	sort.Strings(dirs)
	return dirs[len(dirs)-1], nil
}

func run(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s: %w", filepath.Base(name), err)
	}
	return nil
}
