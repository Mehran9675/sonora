package manager

import (
	"errors"
	"strings"
	"testing"
)

func TestExplain(t *testing.T) {
	err := explain(errors.New("ProxyError: 403 Forbidden"))
	if !strings.HasPrefix(err.Error(), "Couldn't download the model") {
		t.Fatalf("got %q", err)
	}
	if !strings.Contains(err.Error(), "ProxyError") {
		t.Fatal("the original message must be kept")
	}
	plain := errors.New("something else")
	if explain(plain) != plain {
		t.Fatal("unknown errors pass through")
	}
}
