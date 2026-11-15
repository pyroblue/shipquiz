package quiz

import (
	"context"
	"errors"
	"strings"
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
	raw := `{
		"size":"small",
		"summary":"Moves the check.",
		"questions":[{
			"question":"Who owns this boundary?",
			"choices":["The CLI","The API","The cache"],
			"correct_choice":2,
			"explanation":"The API owns the boundary."
		}]
	}`
	got, err := Generate(context.Background(), fakeClient{text: raw}, "model", Diff{Text: "+line\n"}, "default", 1, 7)
	if err != nil {
		t.Fatal(err)
	}
	if got.Questions[0].ID != "q1" || got.Summary == "" {
		t.Fatalf("%#v", got)
	}
}

func TestGenerateKeepsInvalidModelResponse(t *testing.T) {
	raw := `{"questions":[{"id":"q1","question":"Who?","choices":["only one"],"correct_choice":4}]}`
	_, err := Generate(context.Background(), fakeClient{text: raw}, "model", Diff{Text: "+line\n"}, "default", 1, 5)
	var out *ModelOutput
	if !errors.As(err, &out) {
		t.Fatalf("expected model output, got %v", err)
	}
	text := out.Error()
	if !strings.Contains(text, "expected 3 choices, got 1") || !strings.Contains(text, "correct_choice must be 1, 2, or 3") || !strings.Contains(text, raw) {
		t.Fatal(text)
	}
}

func TestGenerateAcceptsZeroBasedCorrectChoice(t *testing.T) {
	raw := `{
		"size":"small",
		"summary":"Logs missing Action env.",
		"questions":[{
			"id":"q2",
			"question":"What should operators always see?",
			"choices":["Provider and model","Only success","Raw config"],
			"correct_choice":0,
			"explanation":"The Action logs the effective target."
		}]
	}`
	got, err := Generate(context.Background(), fakeClient{text: raw}, "model", Diff{Text: "+line\n"}, "default", 1, 5)
	if err != nil {
		t.Fatal(err)
	}
	if got.Questions[0].CorrectChoice < 1 || got.Questions[0].CorrectChoice > 3 {
		t.Fatalf("correct_choice = %d", got.Questions[0].CorrectChoice)
	}
	if got.Questions[0].Choices[got.Questions[0].CorrectChoice-1] != "Provider and model" {
		t.Fatalf("question = %#v", got.Questions[0])
	}
}

type sequenceClient struct {
	replies []string
	calls   int
}

func (s *sequenceClient) Complete(context.Context, llm.Request) (string, error) {
	if s.calls >= len(s.replies) {
		return "", errors.New("no more replies")
	}
	text := s.replies[s.calls]
	s.calls++
	return text, nil
}

func (s *sequenceClient) ListModels(context.Context) ([]string, error) {
	return []string{"model"}, nil
}

func TestGenerateRetriesInvalidQuizOnce(t *testing.T) {
	client := &sequenceClient{replies: []string{
		`{"questions":[{"id":"q1","question":"Who?","choices":["A","B"],"correct_choice":1}]}`,
		`{"size":"tiny","summary":"Fixed.","questions":[{"id":"q1","question":"Who owns it?","choices":["CLI","API","Cache"],"correct_choice":2,"explanation":"The API owns it."}]}`,
	}}
	got, err := Generate(context.Background(), client, "model", Diff{Text: "+line\n"}, "default", 1, 5)
	if err != nil {
		t.Fatal(err)
	}
	if client.calls != 2 || got.Questions[0].Choices[got.Questions[0].CorrectChoice-1] != "API" {
		t.Fatalf("calls=%d got=%#v", client.calls, got)
	}
}
