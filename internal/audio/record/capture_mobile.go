//go:build android || ios

package record

import "errors"

// Voice cloning needs a local model, which phones don't run, so there
// is nothing to record for there.
func open(push func([]byte)) (capture, error) {
	return nil, errors.New("recording isn't available on this device")
}
