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
		"Do not ask about the rest of the application.",
		"answerable from this diff alone",
		"If a candidate is about the application rather than this change, omit it.",
		"Rank those changes by user impact, security, reliability",
		"roughly 18 words or fewer",
		"Diff size alone never justifies more questions.",
		"Give exactly three short, mutually exclusive choices.",
		"correct_choice is 1-based",
		"Never use 0",
	} {
		if !strings.Contains(prompt, want) {
			t.Fatalf("prompt missing %q", want)
		}
	}
	if strings.Contains(prompt, "at most 72 characters") {
		t.Fatal("prompt still forces compressed 72-character questions")
	}
}
