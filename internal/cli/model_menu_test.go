package cli

import "testing"

func TestModelChoicesPinsDefaultAndDropsDatedAlias(t *testing.T) {
	live := []string{
		"claude-3-haiku-20240307",
		"claude-sonnet-4-5",
		"claude-sonnet-4-5-20250929",
		"claude-haiku-4-5",
	}
	got := modelChoices("anthropic", live, "claude-sonnet-4-5")
	if got[0] != "claude-sonnet-4-5" {
		t.Fatalf("first = %q", got[0])
	}
	for _, id := range got {
		if id == "claude-sonnet-4-5-20250929" {
			t.Fatalf("dated alias kept: %#v", got)
		}
	}
	if got[1] != "claude-haiku-4-5" && !contains(got, "claude-haiku-4-5") {
		t.Fatalf("%#v", got)
	}
}

func TestModelChoicesKeepsFlagshipPastAlphabeticalCut(t *testing.T) {
	live := make([]string, 0, 21)
	for i := 0; i < 20; i++ {
		live = append(live, "claude-3-haiku-"+string(rune('a'+i)))
	}
	live = append(live, "claude-sonnet-4-5")
	got := modelChoices("anthropic", live, "claude-haiku-4-5")
	if len(got) > 15 {
		t.Fatalf("len = %d", len(got))
	}
	if got[0] != "claude-haiku-4-5" || !contains(got, "claude-sonnet-4-5") {
		t.Fatalf("%#v", got)
	}
}

func TestModelChoicesFallsBackWhenProviderListsNothing(t *testing.T) {
	got := modelChoices("openai", nil, "gpt-4.1-mini")
	if len(got) != 1 || got[0] != "gpt-4.1-mini" {
		t.Fatalf("%#v", got)
	}
}

func contains(ids []string, want string) bool {
	for _, id := range ids {
		if id == want {
			return true
		}
	}
	return false
}
