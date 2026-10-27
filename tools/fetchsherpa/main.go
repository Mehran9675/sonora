// Command fetchsherpa downloads the sherpa-onnx native libraries that
// the on-device engines link against into third_party/sherpa-onnx.
//
//	go run ./tools/fetchsherpa android     # all Android ABIs, for the APK
//	go run ./tools/fetchsherpa linux-x64   # for developing on Linux
//
// Android builds use them automatically (see tools/packapk); desktop
// builds need -tags sherpa. See internal/engines/onnx/sherpa.
package main

import (
	"archive/tar"
	"compress/bzip2"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// Version is the sherpa-onnx release the binding's header matches.
const Version = "1.13.8"

// libs are the files each target needs.
var libs = []string{"libsherpa-onnx-c-api.so", "libonnxruntime.so"}

func main() {
	if len(os.Args) != 2 || (os.Args[1] != "android" && os.Args[1] != "linux-x64") {
		log.Fatal("usage: fetchsherpa android|linux-x64")
	}
	target := os.Args[1]
	asset := "sherpa-onnx-v" + Version + "-android.tar.bz2"
	if target == "linux-x64" {
		asset = "sherpa-onnx-v" + Version + "-linux-x64-shared.tar.bz2"
	}
	url := "https://github.com/k2-fsa/sherpa-onnx/releases/download/v" + Version + "/" + asset
	out := filepath.Join("third_party", "sherpa-onnx", target)
	log.Printf("downloading %s", url)
	resp, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		log.Fatalf("%s: %s", url, resp.Status)
	}
	tr := tar.NewReader(bzip2.NewReader(resp.Body))
	n := 0
	for {
		h, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		name := path.Base(h.Name)
		if h.Typeflag != tar.TypeReg || !wanted(name) {
			continue
		}
		// Android archives hold jniLibs/<abi>/lib*.so; keep the ABI.
		dir := out
		if target == "android" {
			dir = filepath.Join(out, path.Base(path.Dir(h.Name)))
			if strings.HasPrefix(path.Base(path.Dir(h.Name)), "x86") && path.Base(path.Dir(h.Name)) != "x86_64" {
				continue // 32-bit x86 isn't built
			}
		}
		if err := os.MkdirAll(dir, 0o755); err != nil {
			log.Fatal(err)
		}
		f, err := os.Create(filepath.Join(dir, name))
		if err != nil {
			log.Fatal(err)
		}
		if _, err := io.Copy(f, tr); err != nil {
			log.Fatal(err)
		}
		f.Close()
		n++
	}
	if n == 0 {
		log.Fatal("no libraries found in the archive")
	}
	fmt.Printf("%d libraries in %s\n", n, out)
}

func wanted(name string) bool {
	for _, l := range libs {
		if name == l {
			return true
		}
	}
	return false
}
