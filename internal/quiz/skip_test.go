package quiz

import "testing"

func TestShouldSkip(t *testing.T) {
	if !ShouldSkip([]string{"go.sum", "docs/readme.md"}, []string{"docs/**", "*.md"}) {
		t.Fatal("expected lockfile and docs to skip")
	}
	if ShouldSkip([]string{"go.sum", "main.go"}, []string{"docs/**"}) {
		t.Fatal("a source file should keep the quiz")
	}
	if ShouldSkip(nil, nil) {
		t.Fatal("empty file list is not a skip")
	}
}

func TestMatchPath(t *testing.T) {
	cases := []struct {
		pattern string
		path    string
		want    bool
	}{
		{"*.md", "README.md", true},
		{"*.md", "docs/guide.md", true},
		{"docs/**", "docs/guide.md", true},
		{"docs/**", "src/docs.go", false},
		{"**/*_test.go", "internal/quiz/skip_test.go", true},
		{"**/*_test.go", "internal/quiz/skip.go", false},
	}
	for _, tc := range cases {
		if got := MatchPath(tc.pattern, tc.path); got != tc.want {
			t.Errorf("MatchPath(%q, %q) = %v, want %v", tc.pattern, tc.path, got, tc.want)
		}
	}
}

func TestParsePushRefs(t *testing.T) {
	raw := "refs/heads/main 0123456789abcdef0123456789abcdef01234567 refs/heads/main 0000000000000000000000000000000000000000\n"
	refs := ParsePushRefs(raw)
	if len(refs) != 1 {
		t.Fatalf("refs = %#v", refs)
	}
	if !IsZeroSHA(refs[0].RemoteSHA) || IsZeroSHA(refs[0].LocalSHA) {
		t.Fatalf("sha classification failed: %#v", refs[0])
	}
}

func TestSubstantiveLines(t *testing.T) {
	diff := "--- a/a.go\n+++ b/a.go\n@@\n-old\n+new\n+also\n context\n"
	if got := SubstantiveLines(diff); got != 3 {
		t.Fatalf("lines = %d", got)
	}
}
