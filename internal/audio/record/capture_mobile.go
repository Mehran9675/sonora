//go:build ios

package record

import "errors"

// iOS needs an audio session and a usage description that the build
// doesn't provide yet.
func open(push func([]byte)) (capture, error) {
	return nil, errors.New("recording isn't available on this device")
}
