package quiz

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestWorkingDiffIncludesUncommittedChange(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("GIT_AUTHOR_NAME", "shipquiz")
	t.Setenv("GIT_AUTHOR_EMAIL", "test@shipquiz.test")
	t.Setenv("GIT_COMMITTER_NAME", "shipquiz")
	t.Setenv("GIT_COMMITTER_EMAIL", "test@shipquiz.test")
	git(t, dir, "init", "-b", "main")
	if err := os.WriteFile(filepath.Join(dir, "a.txt"), []byte("one\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	git(t, dir, "add", "a.txt")
	git(t, dir, "commit", "-m", "init")
	if err := os.WriteFile(filepath.Join(dir, "a.txt"), []byte("one\ntwo\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Chdir(dir)

	diff, err := WorkingDiff()
	if err != nil {
		t.Fatal(err)
	}
	if diff.Empty() {
		t.Fatal("expected a diff")
	}
	if len(diff.Files) != 1 || diff.Files[0] != "a.txt" {
		t.Fatalf("files = %#v", diff.Files)
	}
	if SubstantiveLines(diff.Text) < 1 {
		t.Fatalf("diff = %s", diff.Text)
	}
}

func git(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = os.Environ()
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}
