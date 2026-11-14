package quiz

import "testing"

func TestParseGeneratedStripsFence(t *testing.T) {
	raw := "```json\n{\"size\":\"tiny\",\"summary\":\"Adds a cache.\",\"questions\":[{\"id\":\"q1\",\"question\":\"What happens on a miss?\",\"why_it_matters\":\"A miss can stampede.\"}]}\n```"
	g, err := ParseGenerated(raw)
	if err != nil {
		t.Fatal(err)
	}
	if g.Size != "tiny" || len(g.Questions) != 1 || g.Questions[0].ID != "q1" {
		t.Fatalf("%#v", g)
	}
}

func TestNormalizeReplyDropsEmptyAttempts(t *testing.T) {
	questions := []Question{{ID: "q1"}, {ID: "q2"}}
	got := NormalizeReply(Reply{
		Attempted: true,
		Answers:   []Answer{{ID: "q2", Text: "  the lock is process-local  "}},
	}, questions)
	if !got.Attempted || got.Answers[0].Text != "" || got.Answers[1].Text != "the lock is process-local" {
		t.Fatalf("%#v", got)
	}
	empty := NormalizeReply(Reply{Attempted: true, Answers: []Answer{{ID: "q1", Text: "  "}}}, questions)
	if empty.Attempted {
		t.Fatal("expected no attempt")
	}
}
