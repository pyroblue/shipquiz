package cli

import (
	"testing"

	"github.com/shipquiz/shipquiz/internal/quiz"
)

func TestMissedQuestionsIncludesPartialAndWrong(t *testing.T) {
	questions := []quiz.Question{
		{ID: "q1"},
		{ID: "q2"},
		{ID: "q3"},
	}
	ev := quiz.Evaluation{PerQuestion: []quiz.PerQuestion{
		{ID: "q1", Score: 1},
		{ID: "q2", Score: 2},
		{ID: "q3", Score: 0},
	}}

	got := missedQuestions(questions, ev)
	if len(got) != 2 || got[0].ID != "q1" || got[1].ID != "q3" {
		t.Fatalf("missed = %#v", got)
	}
}

func TestMergeAnswersReplacesOnlyRetriedAnswers(t *testing.T) {
	current := []quiz.Answer{
		{ID: "q1", Text: "old one"},
		{ID: "q2", Text: "keep two"},
		{ID: "q3", Text: "old three"},
	}
	retried := []quiz.Answer{
		{ID: "q1", Text: "new one"},
		{ID: "q3", Text: "new three"},
	}

	got := mergeAnswers(current, retried)
	if got[0].Text != "new one" || got[1].Text != "keep two" || got[2].Text != "new three" {
		t.Fatalf("answers = %#v", got)
	}
}
