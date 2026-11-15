package action

import (
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/shipquiz/shipquiz/internal/config"
)

func TestApplyActionEnvLogsMissingProviderAndModel(t *testing.T) {
	t.Setenv("SHIPQUIZ_PROVIDER", "")
	t.Setenv("SHIPQUIZ_MODEL", "")
	t.Setenv("SHIPQUIZ_BASE_URL", "")
	stderr := captureFile(t, &os.Stderr)

	cfg := applyActionEnv(config.Config{Provider: "anthropic", Model: "claude-sonnet-4-5"})
	got := stderr.stop()
	if cfg.Provider != "anthropic" || cfg.Model != "claude-sonnet-4-5" {
		t.Fatalf("cfg = %#v", cfg)
	}
	for _, want := range []string{
		"SHIPQUIZ_PROVIDER is not set; using anthropic",
		"SHIPQUIZ_MODEL is not set; using claude-sonnet-4-5",
		"SHIPQUIZ_BASE_URL is not set; using the provider default",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("stderr missing %q:\n%s", want, got)
		}
	}
}

func TestApplyActionEnvKeepsExplicitValuesQuiet(t *testing.T) {
	t.Setenv("SHIPQUIZ_PROVIDER", "xai")
	t.Setenv("SHIPQUIZ_MODEL", "grok-4.6")
	t.Setenv("SHIPQUIZ_BASE_URL", "https://api.x.ai/v1")
	stderr := captureFile(t, &os.Stderr)

	cfg := applyActionEnv(config.Config{Provider: "anthropic", Model: "claude-sonnet-4-5"})
	got := stderr.stop()
	if cfg.Provider != "xai" || cfg.Model != "grok-4.6" || cfg.BaseURL != "https://api.x.ai/v1" {
		t.Fatalf("cfg = %#v", cfg)
	}
	if strings.Contains(got, "is not set") {
		t.Fatalf("stderr = %q", got)
	}
}

// TestMissingAPIKeyWarnsAndContinues checks that a pull request without the
// model key skips the quiz instead of failing the job.
func TestMissingAPIKeyWarnsAndContinues(t *testing.T) {
	dir := t.TempDir()
	eventPath := filepath.Join(dir, "event.json")
	if err := os.WriteFile(eventPath, []byte(`{"pull_request":{"number":1}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("GITHUB_EVENT_PATH", eventPath)
	t.Setenv("GITHUB_TOKEN", "token")
	t.Setenv("GITHUB_REPOSITORY", "pyroblue/shipquiz")
	t.Setenv("GITHUB_EVENT_NAME", "pull_request")
	t.Setenv("SHIPQUIZ_LLM_API_KEY", "  ")
	t.Setenv("SHIPQUIZ_CONFIG_DIR", filepath.Join(dir, "config"))
	t.Setenv("GITHUB_ACTIONS", "true")

	stderr := captureFile(t, &os.Stderr)
	stdout := captureFile(t, &os.Stdout)
	if code := Run(context.Background()); code != 0 {
		t.Fatalf("exit = %d, want 0", code)
	}
	if got := stderr.stop(); !strings.Contains(got, "SHIPQUIZ_LLM_API_KEY is not set") {
		t.Fatalf("stderr = %q", got)
	}
	if got := stdout.stop(); !strings.Contains(got, "::warning::") {
		t.Fatalf("stdout = %q", got)
	}
}

// captured replaces a process file until stop puts it back and returns what was written.
type captured struct {
	target   **os.File
	original *os.File
	writer   *os.File
	reader   *os.File
}

// captureFile sends writes to target into a pipe for the rest of the test.
func captureFile(t *testing.T, target **os.File) *captured {
	t.Helper()
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	got := &captured{target: target, original: *target, writer: writer, reader: reader}
	*target = writer
	t.Cleanup(func() { got.stop() })
	return got
}

// stop restores the original file and returns everything written since capture.
func (c *captured) stop() string {
	if c.writer == nil {
		return ""
	}
	_ = c.writer.Close()
	c.writer = nil
	*c.target = c.original
	var buf bytes.Buffer
	_, _ = io.Copy(&buf, c.reader)
	_ = c.reader.Close()
	return buf.String()
}
