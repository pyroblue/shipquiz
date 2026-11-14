// Package quiz turns a git diff into questions, collects answers, and grades them.
package quiz

import (
	"fmt"
	"strings"

	"github.com/shipquiz/shipquiz/internal/gitx"
)

// Diff is the change shipquiz will ask about.
type Diff struct {
	Text  string
	Base  string
	Head  string
	Files []string
	Stat  string
}

// Empty reports whether there is nothing substantive to quiz.
func (d Diff) Empty() bool {
	return strings.TrimSpace(d.Text) == ""
}

// Ref is one pre-push hook stdin record.
type Ref struct {
	LocalRef  string
	LocalSHA  string
	RemoteRef string
	RemoteSHA string
}

// ParsePushRefs parses the lines git writes to a pre-push hook.
func ParsePushRefs(input string) []Ref {
	var refs []Ref
	for _, line := range strings.Split(input, "\n") {
		fields := strings.Fields(line)
		if len(fields) < 4 {
			continue
		}
		refs = append(refs, Ref{
			LocalRef:  fields[0],
			LocalSHA:  fields[1],
			RemoteRef: fields[2],
			RemoteSHA: fields[3],
		})
	}
	return refs
}

// IsZeroSHA reports whether sha is git's all-zero "missing object" id.
func IsZeroSHA(sha string) bool {
	if sha == "" {
		return true
	}
	for _, c := range sha {
		if c != '0' {
			return false
		}
	}
	return true
}

// WorkingDiff is the manual-quiz diff: uncommitted changes, otherwise
// the unpushed range against upstream, otherwise the last commit.
func WorkingDiff() (Diff, error) {
	work, err := gitx.Output("diff", "HEAD")
	if err != nil {
		return Diff{}, err
	}
	if strings.TrimSpace(work) != "" {
		return readDiff("HEAD", "worktree", "diff", "HEAD")
	}
	if _, err := gitx.Output("rev-parse", "--abbrev-ref", "--symbolic-full-name", "@{upstream}"); err == nil {
		return readDiff("@{upstream}", "HEAD", "diff", "@{upstream}...HEAD")
	}
	parent, err := gitx.Output("rev-parse", "--verify", "HEAD~1")
	if err != nil {
		return readDiff("", "HEAD", "show", "--format=", "HEAD")
	}
	base := strings.TrimSpace(parent)
	return readDiff(base, "HEAD", "diff", base+"...HEAD")
}

// PushDiff builds the diff for every ref git is about to push.
func PushDiff(refs []Ref) (Diff, error) {
	if len(refs) == 0 {
		return WorkingDiff()
	}
	var parts []string
	var files []string
	var stats []string
	base := ""
	head := ""
	for _, ref := range refs {
		if IsZeroSHA(ref.LocalSHA) {
			continue
		}
		from := ref.RemoteSHA
		if IsZeroSHA(from) {
			resolved, err := baseForNewBranch(ref.LocalSHA)
			if err != nil {
				return Diff{}, err
			}
			from = resolved
		}
		piece, err := readDiff(from, ref.LocalSHA, "diff", from+"..."+ref.LocalSHA)
		if err != nil {
			return Diff{}, err
		}
		if piece.Empty() {
			continue
		}
		parts = append(parts, piece.Text)
		files = append(files, piece.Files...)
		stats = append(stats, piece.Stat)
		base = from
		head = ref.LocalSHA
	}
	if len(parts) == 0 {
		return Diff{}, nil
	}
	return Diff{
		Text:  truncate(strings.Join(parts, "\n")),
		Base:  base,
		Head:  head,
		Files: unique(files),
		Stat:  strings.TrimSpace(strings.Join(stats, "\n")),
	}, nil
}

// RangeDiff is the diff between two commits, used by the GitHub Action.
func RangeDiff(base, head string) (Diff, error) {
	return readDiff(base, head, "diff", base+"..."+head)
}

func readDiff(base, head string, gitArgs ...string) (Diff, error) {
	text, err := gitx.Output(gitArgs...)
	if err != nil {
		return Diff{}, err
	}
	nameArgs := append([]string{"diff", "--name-only"}, gitArgs[1:]...)
	if gitArgs[0] == "show" {
		nameArgs = []string{"show", "--name-only", "--format=", "HEAD"}
	}
	names, err := gitx.Output(nameArgs...)
	if err != nil {
		names = ""
	}
	statArgs := append([]string{"diff", "--stat"}, gitArgs[1:]...)
	if gitArgs[0] == "show" {
		statArgs = []string{"show", "--stat", "--format=", "HEAD"}
	}
	stat, err := gitx.Output(statArgs...)
	if err != nil {
		stat = ""
	}
	return Diff{
		Text:  truncate(text),
		Base:  base,
		Head:  head,
		Files: splitLines(names),
		Stat:  strings.TrimSpace(stat),
	}, nil
}

func baseForNewBranch(head string) (string, error) {
	candidates := []string{}
	if out, err := gitx.Output("symbolic-ref", "--short", "refs/remotes/origin/HEAD"); err == nil {
		candidates = append(candidates, strings.TrimSpace(out))
	}
	candidates = append(candidates, "origin/main", "origin/master", "main", "master")
	seen := map[string]struct{}{}
	for _, candidate := range candidates {
		if _, ok := seen[candidate]; ok {
			continue
		}
		seen[candidate] = struct{}{}
		out, err := gitx.Output("merge-base", head, candidate)
		if err == nil && strings.TrimSpace(out) != "" {
			return strings.TrimSpace(out), nil
		}
	}
	out, err := gitx.Output("rev-parse", head+"^")
	if err == nil && strings.TrimSpace(out) != "" {
		return strings.TrimSpace(out), nil
	}
	return "", fmt.Errorf("could not find a base commit for %s", head)
}

func splitLines(raw string) []string {
	var lines []string
	for _, line := range strings.Split(raw, "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			lines = append(lines, line)
		}
	}
	return unique(lines)
}

func unique(items []string) []string {
	seen := map[string]struct{}{}
	var out []string
	for _, item := range items {
		if _, ok := seen[item]; ok {
			continue
		}
		seen[item] = struct{}{}
		out = append(out, item)
	}
	return out
}

const maxDiffBytes = 48_000

func truncate(text string) string {
	if len(text) <= maxDiffBytes {
		return text
	}
	return text[:maxDiffBytes] + "\n\n[diff truncated]\n"
}
