package hook

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestInstallPreservesExistingHook(t *testing.T) {
	dir := t.TempDir()
	git(t, dir, "init", "-b", "main")
	hooks := filepath.Join(dir, ".git", "hooks")
	if err := os.MkdirAll(hooks, 0o755); err != nil {
		t.Fatal(err)
	}
	previous := "#!/bin/sh\necho old\n"
	if err := os.WriteFile(filepath.Join(hooks, "pre-push"), []byte(previous), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Chdir(dir)

	if err := Install(); err != nil {
		t.Fatal(err)
	}
	if !Installed() {
		t.Fatal("expected hook to be installed")
	}
	got, err := os.ReadFile(filepath.Join(hooks, "pre-push.previous"))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != previous {
		t.Fatalf("previous hook = %q", got)
	}
	script, err := os.ReadFile(filepath.Join(hooks, "pre-push"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(script), "shipquiz hook-run") {
		t.Fatal("hook does not call shipquiz")
	}

	if err := Uninstall(); err != nil {
		t.Fatal(err)
	}
	if Installed() {
		t.Fatal("hook still installed")
	}
	restored, err := os.ReadFile(filepath.Join(hooks, "pre-push"))
	if err != nil {
		t.Fatal(err)
	}
	if string(restored) != previous {
		t.Fatalf("restored = %q", restored)
	}
}

func git(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}
