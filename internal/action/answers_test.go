package action

import (
	"strings"
	"testing"

	"github.com/shipquiz/shipquiz/internal/quiz"
)

func TestParseChoiceReply(t *testing.T) {
	questions := []quiz.Question{
		{ID: "q1", Choices: []string{"A", "B", "C"}},
		{ID: "q2", Choices: []string{"D", "E", "F"}},
		{ID: "q3", Choices: []string{"G", "H", "I"}},
	}
	got, ok := parseChoiceReply("3, 3, 2", questions)
	if !ok || !got.Attempted || len(got.Answers) != 3 {
		t.Fatalf("got = %#v ok=%v", got, ok)
	}
	if got.Answers[0].Choice != 3 || got.Answers[0].Text != "C" {
		t.Fatalf("q1 = %#v", got.Answers[0])
	}
	if got.Answers[2].Choice != 2 || got.Answers[2].Text != "H" {
		t.Fatalf("q3 = %#v", got.Answers[2])
	}
	if _, ok := parseChoiceReply("LGTM", questions); ok {
		t.Fatal("accepted a non-answer")
	}
}

func TestMarkerRoundTrip(t *testing.T) {
	questions := []quiz.Question{
		{
			ID:            "q1",
			Question:      "Who owns this boundary?",
			Reference:     "The API owns the boundary.",
			Choices:       []string{"The CLI", "The API", "The cache"},
			CorrectChoice: 2,
			Explanation:   "The API owns the boundary.",
		},
		{ID: "q2", Question: "What fails closed?"},
	}
	marker := Marker{V: 1, HeadSHA: "abc", CheckRunID: 9, Quiz: quiz.Generated{
		Size:      "small",
		Summary:   "Moves the boundary check.",
		Questions: questions,
	}}
	rendered, err := renderQuiz(marker)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"**Change Size:** Small",
		"2 questions",
		"**Diff Summary:**",
		"### Q1",
		"2. The API",
	} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("comment missing %q:\n%s", want, rendered)
		}
	}
	if strings.Contains(rendered, "q1:") {
		t.Fatal("quiz comment still requires labeled answers")
	}
	if strings.Contains(rendered, "The API owns the boundary.") {
		t.Fatal("quiz comment includes the answer key")
	}
	got, ok := decodeMarker(rendered)
	if !ok || got.HeadSHA != "abc" || got.CheckRunID != 9 || len(got.Quiz.Questions) != 2 {
		t.Fatalf("marker = %#v ok=%v", got, ok)
	}
	if got.Quiz.Questions[0].CorrectChoice != 0 || got.Quiz.Questions[0].Explanation != "" {
		t.Fatal("hidden marker leaks the answer key")
	}
	result := renderResult(quiz.Evaluation{Overall: "pass", Score: 2, Feedback: "clear"})
	if !strings.Contains(result, "<!-- shipquiz result -->") {
		t.Fatal("result comment would retrigger the workflow")
	}
	if _, ok := decodeMarker(result); ok {
		t.Fatal("result comment decoded as a quiz marker")
	}
}
