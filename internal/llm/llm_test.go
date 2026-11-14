package llm

import "testing"

func TestNewXAIUsesOpenAICompatibleHost(t *testing.T) {
	client, err := New("xai", "", "test-key")
	if err != nil {
		t.Fatal(err)
	}
	got, ok := client.(*openAIClient)
	if !ok {
		t.Fatalf("client type %T", client)
	}
	if got.base != "https://api.x.ai/v1" {
		t.Fatalf("base = %s", got.base)
	}
}
