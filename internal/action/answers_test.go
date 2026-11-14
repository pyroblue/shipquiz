package action

import (
	"strings"
	"testing"

	"github.com/shipquiz/shipquiz/internal/quiz"
)

func TestParseAnswersAndMarker(t *testing.T) {
	questions := []quiz.Question{
		{ID: "q1", Question: "Why this boundary?"},
		{ID: "q2", Question: "What fails closed?"},
	}
	body := "q1: the check moved into the handler\nso the middleware no longer sees it\n\n2. a missing call site returns 401\n"
	answers := ParseAnswers(body, questions)
	if answers[0].Text != "the check moved into the handler\nso the middleware no longer sees it" {
		t.Fatalf("q1 = %q", answers[0].Text)
	}
	if !strings.Contains(answers[1].Text, "401") {
		t.Fatalf("q2 = %q", answers[1].Text)
	}

	marker := Marker{V: 1, HeadSHA: "abc", CheckRunID: 9, Quiz: quiz.Generated{Size: "small", Questions: questions}}
	rendered, err := renderQuiz(marker)
	if err != nil {
		t.Fatal(err)
	}
	got, ok := decodeMarker(rendered)
	if !ok || got.HeadSHA != "abc" || got.CheckRunID != 9 || len(got.Quiz.Questions) != 2 {
		t.Fatalf("marker = %#v ok=%v", got, ok)
	}
}
