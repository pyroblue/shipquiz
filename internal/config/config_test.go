package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSaveLoadAndRepoOverride(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("SHIPQUIZ_CONFIG_DIR", dir)

	cfg := Default()
	cfg.Model = "claude-haiku-4-5"
	cfg.Hooks.PrePush = false
	if err := Save(cfg); err != nil {
		t.Fatal(err)
	}

	loaded, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Model != "claude-haiku-4-5" {
		t.Fatalf("model = %q", loaded.Model)
	}
	if loaded.Hooks.PrePush {
		t.Fatal("pre_push was saved false and loaded true")
	}
	if loaded.Quiz.PassingAverage != 1.5 {
		t.Fatalf("passing average = %v", loaded.Quiz.PassingAverage)
	}

	repo := t.TempDir()
	body := []byte("model = \"claude-sonnet-4-5\"\nskip_paths = [\"docs/**\"]\nquestion_style = \"failure-modes\"\n")
	if err := os.WriteFile(filepath.Join(repo, ".shipquiz.toml"), body, 0o644); err != nil {
		t.Fatal(err)
	}
	merged, err := LoadForRepo(repo)
	if err != nil {
		t.Fatal(err)
	}
	if merged.Model != "claude-sonnet-4-5" {
		t.Fatalf("merged model = %q", merged.Model)
	}
	if merged.QuestionStyle != "failure-modes" {
		t.Fatalf("style = %q", merged.QuestionStyle)
	}
	if len(merged.SkipPaths) != 1 || merged.SkipPaths[0] != "docs/**" {
		t.Fatalf("skip paths = %#v", merged.SkipPaths)
	}
	if merged.Hooks.PrePush {
		t.Fatal("repo file should not clobber hooks.pre_push")
	}
}

func TestSet(t *testing.T) {
	cfg := Default()
	if err := Set(&cfg, "quiz.timeout_seconds", "15"); err != nil {
		t.Fatal(err)
	}
	if cfg.Quiz.TimeoutSeconds != 15 {
		t.Fatalf("timeout = %d", cfg.Quiz.TimeoutSeconds)
	}
	if err := Set(&cfg, "nope", "1"); err == nil {
		t.Fatal("expected unknown key error")
	}
}

func TestMissingFileKeepsPrePushOn(t *testing.T) {
	t.Setenv("SHIPQUIZ_CONFIG_DIR", t.TempDir())
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if !cfg.Hooks.PrePush {
		t.Fatal("default pre_push should be true")
	}
}
