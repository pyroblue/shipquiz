package cli

import "testing"

func TestGitHubSlug(t *testing.T) {
	cases := []struct {
		remote string
		want   string
	}{
		{"git@github.com:pyroblue/shipquiz.git", "pyroblue/shipquiz"},
		{"https://github.com/pyroblue/shipquiz.git", "pyroblue/shipquiz"},
		{"https://github.com/pyroblue/shipquiz", "pyroblue/shipquiz"},
		{"ssh://git@github.com/pyroblue/shipquiz.git", "pyroblue/shipquiz"},
		{"https://gitlab.com/a/b.git", ""},
	}
	for _, tc := range cases {
		if got := githubSlug(tc.remote); got != tc.want {
			t.Fatalf("githubSlug(%q) = %q, want %q", tc.remote, got, tc.want)
		}
	}
}
