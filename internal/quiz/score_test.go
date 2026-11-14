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
		},
	}, 1, 2)
	if len(g.Questions) != 2 {
		t.Fatalf("len = %d", len(g.Questions))
	}
	if g.Questions[0].ID != "q1" || g.Questions[0].Question != "one" {
		t.Fatalf("first = %#v", g.Questions[0])
	}
}
