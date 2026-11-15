package quiz

import "testing"

func TestEnforcePassAndFail(t *testing.T) {
	questions := []Question{{ID: "q1"}, {ID: "q2"}}
	pass := Enforce(Evaluation{
		Overall: "fail",
		PerQuestion: []PerQuestion{
			{ID: "q1", Score: 2},
			{ID: "q2", Score: 1},
		},
	}, questions, 1.5)
	if !pass.Passed() || pass.Score != 1.5 {
		t.Fatalf("expected pass at 1.5, got %#v", pass)
	}

	fail := Enforce(Evaluation{
		Overall: "pass",
		PerQuestion: []PerQuestion{
			{ID: "q1", Score: 2},
			{ID: "q2", Score: 0},
		},
	}, questions, 1.5)
	if fail.Passed() {
		t.Fatal("a zero must fail even when the model says pass")
	}

	missing := Enforce(Evaluation{}, questions, 1.5)
	if missing.Passed() || missing.PerQuestion[1].Score != 0 {
		t.Fatalf("missing scores should be zeros: %#v", missing)
	}
}

func TestNormalizeAssignsIDsAndCap(t *testing.T) {
	g := Normalize(Generated{
		Questions: []Question{
			{Question: " one "},
			{Question: "two"},
			{Question: "three"},
			{Question: "four"},
			{Question: "five"},
			{Question: "six"},
		},
	}, 1, 9)
	if len(g.Questions) != 5 {
		t.Fatalf("len = %d", len(g.Questions))
	}
	if g.Questions[0].ID != "q1" || g.Questions[0].Question != "one" {
		t.Fatalf("first = %#v", g.Questions[0])
	}
}

func TestGradeChoicesIsDeterministic(t *testing.T) {
	questions := []Question{
		{ID: "q1", Choices: []string{"A", "B", "C"}, CorrectChoice: 2, Explanation: "B owns it."},
		{ID: "q2", Choices: []string{"D", "E", "F"}, CorrectChoice: 1, Explanation: "D contains it."},
	}
	ev := GradeChoices(questions, []Answer{
		{ID: "q1", Choice: 2},
		{ID: "q2", Choice: 3},
	})
	if ev.Passed() || ev.Score != 1 || ev.PerQuestion[0].Score != 2 || ev.PerQuestion[1].Score != 0 {
		t.Fatalf("evaluation = %#v", ev)
	}
	if ev.PerQuestion[1].Comment != "D contains it." {
		t.Fatalf("comment = %q", ev.PerQuestion[1].Comment)
	}
}

func TestFixChoiceNumbersConvertsZeroBasedIndexes(t *testing.T) {
	g := Normalize(Generated{Questions: []Question{
		{
			ID:            "q1",
			Question:      "If Action env vars are unset, what happens?",
			Choices:       []string{"Fail", "Keep file config and warn", "Hardcode silently"},
			CorrectChoice: 1,
		},
		{
			ID:            "q2",
			Question:      "What should operators always see?",
			Choices:       []string{"Provider, model, and base URL", "Only success", "Raw config"},
			CorrectChoice: 0,
		},
	}}, 1, 5)
	if g.Questions[0].CorrectChoice != 2 || g.Questions[1].CorrectChoice != 1 {
		t.Fatalf("questions = %#v", g.Questions)
	}
	if !MultipleChoiceReady(g) {
		t.Fatal(ChoiceIssue(g))
	}
}

func TestRotateChoicesKeepsCorrectAnswer(t *testing.T) {
	g := Generated{Questions: []Question{{
		Question:      "Which boundary owns this?",
		Choices:       []string{"wrong", "correct", "also wrong"},
		CorrectChoice: 2,
	}}}
	RotateChoices(&g)
	q := g.Questions[0]
	if q.Choices[q.CorrectChoice-1] != "correct" {
		t.Fatalf("question = %#v", q)
	}
}
