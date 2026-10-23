package engines

import "testing"

func TestResolveAndSubset(t *testing.T) {
	opts := []Option{
		{ID: "model", Scope: ScopeLoad, Default: "a"},
		{ID: "temperature", Default: 0.7},
		{ID: "url", Scope: ScopeSetup},
	}
	v := Resolve(opts, Values{"temperature": 0.2, "stale": true})
	if v.String("model") != "a" || v.Float("temperature", 0) != 0.2 {
		t.Fatalf("resolve: %v", v)
	}
	if _, ok := v["stale"]; ok {
		t.Fatal("values of removed options must be dropped")
	}
	if load := Subset(opts, v, ScopeLoad); len(load) != 1 || load["model"] != "a" {
		t.Fatalf("subset: %v", load)
	}
}
