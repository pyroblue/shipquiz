package quiz

import (
	"context"
	"testing"

	"github.com/shipquiz/shipquiz/internal/llm"
)

func TestCacheRoundTrip(t *testing.T) {
	t.Setenv("SHIPQUIZ_CACHE_DIR", t.TempDir())
	key := Key("a", "b", "+x", "model", "default")
	want := Entry{Generated: Generated{Size: "tiny", Summary: "s", Questions: []Question{{ID: "q1", Question: "Why?"}}}}
	if err := StoreEntry(key, want); err != nil {
		t.Fatal(err)
	}
	got, ok, err := LoadEntry(key)
	if err != nil || !ok {
		t.Fatalf("load ok=%v err=%v", ok, err)
	}
	if got.Generated.Summary != "s" || got.Generated.Questions[0].ID != "q1" {
		t.Fatalf("%#v", got)
	}
}

type fakeClient struct {
	text string
}

func (f fakeClient) Complete(context.Context, llm.Request) (string, error) {
	return f.text, nil
}

func (f fakeClient) ListModels(context.Context) ([]string, error) {
	return []string{"model"}, nil
}

func TestGenerateNormalizes(t *testing.T) {
	raw := `{"size":"small","summary":"Moves the check.","questions":[{"question":"Why here?"}]}`
	got, err := Generate(context.Background(), fakeClient{text: raw}, "model", Diff{Text: "+line\n"}, "default", 1, 7)
	if err != nil {
		t.Fatal(err)
	}
	if got.Questions[0].ID != "q1" || got.Summary == "" {
		t.Fatalf("%#v", got)
	}
}
