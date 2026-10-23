package catalog

import (
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/mehran9675/sonora/internal/engines"
)

// TestBuiltinEngines checks every embedded engine definition: it must
// parse, ship an adapter, and declare options the UI can render.
func TestBuiltinEngines(t *testing.T) {
	dir := t.TempDir()
	script, err := Unpack(dir)
	if err != nil {
		t.Fatal(err)
	}
	ms, errs := LoadDir(filepath.Join(dir, "engines"))
	for _, err := range errs {
		t.Error(err)
	}
	if len(ms) < 8 {
		t.Fatalf("found %d engines", len(ms))
	}
	for _, m := range ms {
		if m.Name == "" || m.Description == "" || len(m.Install) == 0 {
			t.Errorf("%s: name, description and install recipe are required", m.ID)
		}
		for _, o := range m.Options {
			switch o.Keyboard {
			case "", "url", "number", "email", "phone":
			default:
				t.Errorf("%s.%s: unknown keyboard %q", m.ID, o.ID, o.Keyboard)
			}
			switch o.Type {
			case engines.OptionSelect:
				if len(o.Choices) == 0 {
					t.Errorf("%s.%s: select without choices", m.ID, o.ID)
				}
			case engines.OptionNumber:
				if o.Max <= o.Min {
					t.Errorf("%s.%s: empty range", m.ID, o.ID)
				}
			case engines.OptionText, engines.OptionSecret, engines.OptionBool:
			default:
				t.Errorf("%s.%s: unknown type %q", m.ID, o.ID, o.Type)
			}
		}
	}
	// The Python sources must at least compile.
	py, err := exec.LookPath("python3")
	if err != nil {
		t.Skip("python3 not available")
	}
	if out, err := exec.Command(py, "-m", "compileall", "-q", filepath.Dir(script), filepath.Join(dir, "engines")).CombinedOutput(); err != nil {
		t.Fatalf("python sources don't compile: %v\n%s", err, out)
	}
}
