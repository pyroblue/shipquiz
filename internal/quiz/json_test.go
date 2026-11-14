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
