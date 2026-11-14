package action

import (
	"os"
	"strings"
	"testing"
)

func TestWorkflowUsesLLMSecret(t *testing.T) {
	got := Workflow("openai", "gpt-4.1-mini", "https://example.test/v1")
	for _, want := range []string{
		"SHIPQUIZ_LLM_API_KEY",
		"SHIPQUIZ_PROVIDER: openai",
		"SHIPQUIZ_MODEL: gpt-4.1-mini",
		`SHIPQUIZ_BASE_URL: "https://example.test/v1"`,
		"!contains(github.event.comment.body, '<!-- shipquiz')",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("workflow missing %q", want)
		}
	}
	if strings.Contains(got, "q1:") || strings.Contains(got, "SHIPQUIZ_API_KEY:") {
		t.Fatalf("workflow still has the old reply format or secret name:\n%s", got)
	}
}

func TestCheckedInWorkflowMatchesTemplate(t *testing.T) {
	got, err := os.ReadFile("../../.github/workflows/shipquiz.yml")
	if err != nil {
		t.Fatal(err)
	}
	want := Workflow("anthropic", "claude-sonnet-4-5", "")
	if string(got) != want {
		t.Fatalf("checked-in workflow drifted from the template")
	}
}
