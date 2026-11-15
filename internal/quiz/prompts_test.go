package quiz

import (
	"strings"
	"testing"
)

func TestGenerationPromptUsesArchitectureQualityGates(t *testing.T) {
	_, prompt := GenerationPrompt(Diff{
		Text:  "diff --git a/a.go b/a.go\n+func helper() {}",
		Files: []string{"a.go"},
	}, "default", 1, 5)

	for _, want := range []string{
		"Rank them by user impact, security, reliability",
		"If a candidate does not pass this gate, omit it.",
		"roughly 18 words or fewer",
		"Avoid implementation verbs",
		"Reject questions answered by merely restating the diff",
		"Diff size alone never justifies more questions.",
		"Give exactly three short, mutually exclusive choices.",
		"correct_choice",
	} {
		if !strings.Contains(prompt, want) {
			t.Fatalf("prompt missing %q", want)
		}
	}
	if strings.Contains(prompt, "at most 72 characters") {
		t.Fatal("prompt still forces compressed 72-character questions")
	}
}
