package keychain

import (
	"testing"
)

func TestFileStoreRoundTrip(t *testing.T) {
	t.Setenv("SHIPQUIZ_CONFIG_DIR", t.TempDir())
	t.Setenv("SHIPQUIZ_KEY_STORE", "file")
	t.Setenv("SHIPQUIZ_LLM_API_KEY", "")

	if err := Set("anthropic", "sk-test"); err != nil {
		t.Fatal(err)
	}
	got, err := Get("anthropic")
	if err != nil {
		t.Fatal(err)
	}
	if got != "sk-test" {
		t.Fatalf("got %q", got)
	}
}

func TestEnvOverridesStore(t *testing.T) {
	t.Setenv("SHIPQUIZ_CONFIG_DIR", t.TempDir())
	t.Setenv("SHIPQUIZ_KEY_STORE", "file")
	if err := Set("openai", "from-file"); err != nil {
		t.Fatal(err)
	}
	t.Setenv("SHIPQUIZ_LLM_API_KEY", "from-env")
	got, err := Get("openai")
	if err != nil {
		t.Fatal(err)
	}
	if got != "from-env" {
		t.Fatalf("got %q", got)
	}
}
