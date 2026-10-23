package command

import (
	"reflect"
	"testing"
)

func TestSplitAndExpand(t *testing.T) {
	args, err := splitArgs(`say -v "{voice}" -o '{output}' {text}`)
	if err != nil {
		t.Fatal(err)
	}
	got := expand(args, map[string]string{"voice": "Good News", "output": "/tmp/a b.wav", "text": "hello world"})
	want := []string{"say", "-v", "Good News", "-o", "/tmp/a b.wav", "hello world"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %q want %q", got, want)
	}
	if _, err := splitArgs(`a "b`); err == nil {
		t.Fatal("expected unclosed quote error")
	}
}
