// Package keys finds the text Sonora translates: the keys of tr.T,
// tr.N, tr.Msg and tr.Bind calls in the UI code, the status and hint
// messages the engine manager reports, and the descriptions, option
// labels and language names engines show.
package keys

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/mehran9675/sonora/internal/catalog"
	"github.com/mehran9675/sonora/internal/engines"
	"github.com/mehran9675/sonora/internal/engines/onnx"
)

// Key is a message to translate.
type Key struct {
	Text   string `json:"text"`
	Plural bool   `json:"plural,omitempty"` // used with tr.N
	Where  string `json:"where"`            // the first place it's used
}

// Find returns the keys under the repository root, sorted.
func Find(root string) ([]Key, error) {
	found := map[string]*Key{}
	add := func(text, where string, plural bool) {
		if strings.TrimSpace(text) == "" {
			return
		}
		k := found[text]
		if k == nil {
			k = &Key{Text: text, Where: where}
			found[text] = k
		}
		k.Plural = k.Plural || plural
	}
	if err := goKeys(filepath.Join(root, "ui"), root, add); err != nil {
		return nil, err
	}
	if err := managerKeys(filepath.Join(root, "internal", "manager"), root, add); err != nil {
		return nil, err
	}
	manifests, _ := catalog.LoadDir(filepath.Join(root, "python", "engines"))
	tmp, err := os.MkdirTemp("", "i18nkeys")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(tmp)
	for _, e := range onnx.All(tmp) {
		manifests = append(manifests, e.Manifest())
	}
	for _, m := range manifests {
		manifestKeys(m, add)
	}
	for _, l := range engines.Languages {
		add(l.Label, "engines.Languages", false)
	}
	out := make([]Key, 0, len(found))
	for _, k := range found {
		out = append(out, *k)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Text < out[j].Text })
	return out, nil
}

// goKeys collects the literal keys of tr calls in the Go files under dir.
func goKeys(dir, root string, add func(text, where string, plural bool)) error {
	fset := token.NewFileSet()
	return filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return err
		}
		f, err := parser.ParseFile(fset, path, nil, 0)
		if err != nil {
			return err
		}
		ast.Inspect(f, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}
			sel, ok := call.Fun.(*ast.SelectorExpr)
			if !ok {
				return true
			}
			pkg, ok := sel.X.(*ast.Ident)
			if !ok || pkg.Name != "tr" {
				return true
			}
			arg := 0
			switch sel.Sel.Name {
			case "T", "N", "Msg":
			case "Bind":
				arg = 1
			default:
				return true
			}
			if len(call.Args) <= arg {
				return true
			}
			if s, ok := literal(call.Args[arg]); ok {
				rel, _ := filepath.Rel(root, fset.Position(call.Pos()).Filename)
				add(s, rel, sel.Sel.Name == "N")
			}
			return true
		})
		return nil
	})
}

// managerKeys collects the manager's status messages (string literals
// assigned to a Status field) and its error hints (the text field of
// the hints table).
func managerKeys(dir, root string, add func(text, where string, plural bool)) error {
	fset := token.NewFileSet()
	files, err := filepath.Glob(filepath.Join(dir, "*.go"))
	if err != nil {
		return err
	}
	for _, path := range files {
		if strings.HasSuffix(path, "_test.go") {
			continue
		}
		f, err := parser.ParseFile(fset, path, nil, 0)
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(root, path)
		ast.Inspect(f, func(n ast.Node) bool {
			switch n := n.(type) {
			case *ast.AssignStmt:
				for i, lhs := range n.Lhs {
					if sel, ok := lhs.(*ast.SelectorExpr); ok && sel.Sel.Name == "Status" && i < len(n.Rhs) {
						if s, ok := literal(n.Rhs[i]); ok {
							add(s, rel, false)
						}
					}
				}
			case *ast.ValueSpec:
				for _, name := range n.Names {
					if name.Name != "hints" || len(n.Values) == 0 {
						continue
					}
					table, ok := n.Values[0].(*ast.CompositeLit)
					if !ok {
						continue
					}
					for _, el := range table.Elts {
						if row, ok := el.(*ast.CompositeLit); ok && len(row.Elts) > 1 {
							if s, ok := literal(row.Elts[len(row.Elts)-1]); ok {
								add(s, rel, false)
							}
						}
					}
				}
			}
			return true
		})
	}
	return nil
}

// manifestKeys collects what an engine shows besides names: its
// description, hint and tags, its voices' groups and descriptions, its options' labels, help and placeholders,
// and the labels of choices other than voices and models.
func manifestKeys(m *engines.Manifest, add func(text, where string, plural bool)) {
	where := "engine " + m.ID
	add(m.Description, where, false)
	add(m.Hint, where, false)
	for _, t := range m.Tags {
		add(t, where, false)
	}
	for _, v := range m.Voices {
		// Shown through core.Parts: each " · " part translates alone.
		for _, s := range []string{v.Group, v.Description} {
			for _, part := range strings.Split(s, " · ") {
				if part != "" && (part[0] < '0' || part[0] > '9') {
					add(part, where, false)
				}
			}
		}
	}
	for _, o := range m.Options {
		add(o.Label, where, false)
		add(o.Help, where, false)
		add(o.Placeholder, where, false)
		if o.ID == "voice" || o.ID == "speaker" || o.ID == m.ModelOptionID || (m.ModelOptionID == "" && o.ID == "model") {
			continue
		}
		for _, c := range o.Choices {
			add(c.Label, where, false)
		}
	}
}

// literal returns a string literal's value, or the value of literals
// joined with +.
func literal(e ast.Expr) (string, bool) {
	switch e := e.(type) {
	case *ast.BasicLit:
		if e.Kind != token.STRING {
			return "", false
		}
		s, err := strconv.Unquote(e.Value)
		return s, err == nil
	case *ast.BinaryExpr:
		if e.Op != token.ADD {
			return "", false
		}
		a, ok1 := literal(e.X)
		b, ok2 := literal(e.Y)
		return a + b, ok1 && ok2
	case *ast.ParenExpr:
		return literal(e.X)
	}
	return "", false
}
