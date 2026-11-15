package action

import (
	"strings"
	"testing"

	"github.com/shipquiz/shipquiz/internal/quiz"
)

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
	marker := Marker{V: 1, HeadSHA: "abc", CheckRunID: 9, Quiz: quiz.Generated{Size: "small", Questions: questions}}
	rendered, err := renderQuiz(marker)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(rendered, "q1:") {
		t.Fatal("quiz comment still requires labeled answers")
	}
	if strings.Contains(rendered, "The API owns the boundary.") {
		t.Fatal("quiz comment includes the answer key")
	}
	if !strings.Contains(rendered, "2. The API") {
		t.Fatal("quiz comment omits the choices")
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
