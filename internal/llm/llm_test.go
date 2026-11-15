package llm

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

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
	if got.reasoning != "low" {
		t.Fatalf("reasoning = %q", got.reasoning)
	}
}

func TestXAISendsLowReasoningEffort(t *testing.T) {
	var body map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("decode: %v", err)
		}
		io.WriteString(w, `{"choices":[{"message":{"content":"ok"}}]}`)
	}))
	defer srv.Close()

	client, err := New("xai", srv.URL, "test-key")
	if err != nil {
		t.Fatal(err)
	}
	text, err := client.Complete(context.Background(), Request{Model: "grok-4.6", User: "quiz this diff"})
	if err != nil {
		t.Fatal(err)
	}
	if text != "ok" {
		t.Fatalf("text = %q", text)
	}
	if body["reasoning_effort"] != "low" {
		t.Fatalf("reasoning_effort = %v", body["reasoning_effort"])
	}
}

func TestOpenAICompatibleSendsJSONObjectFormat(t *testing.T) {
	var body map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("decode: %v", err)
		}
		io.WriteString(w, `{"choices":[{"message":{"content":"{}"}}]}`)
	}))
	defer srv.Close()

	client, err := New("openai", srv.URL, "test-key")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := client.Complete(context.Background(), Request{Model: "m", User: "return json", JSONObject: true}); err != nil {
		t.Fatal(err)
	}
	format, _ := body["response_format"].(map[string]any)
	if format["type"] != "json_object" {
		t.Fatalf("response_format = %#v", body["response_format"])
	}
}

func TestOpenAIAndGeminiSendLowReasoningEffort(t *testing.T) {
	for _, provider := range []string{"openai", "gemini"} {
		t.Run(provider, func(t *testing.T) {
			var body map[string]any
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					t.Errorf("decode: %v", err)
				}
				io.WriteString(w, `{"choices":[{"message":{"content":"ok"}}]}`)
			}))
			defer srv.Close()

			client, err := New(provider, srv.URL, "test-key")
			if err != nil {
				t.Fatal(err)
			}
			if _, err := client.Complete(context.Background(), Request{Model: "m", User: "hi"}); err != nil {
				t.Fatal(err)
			}
			if body["reasoning_effort"] != "low" {
				t.Fatalf("reasoning_effort = %v", body["reasoning_effort"])
			}
		})
	}
}

func TestDropsRejectedReasoningEffort(t *testing.T) {
	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("decode: %v", err)
		}
		calls++
		if _, ok := body["reasoning_effort"]; ok {
			http.Error(w, `{"error":{"message":"unknown parameter reasoning_effort"}}`, http.StatusBadRequest)
			return
		}
		io.WriteString(w, `{"choices":[{"message":{"content":"ok"}}]}`)
	}))
	defer srv.Close()

	client, err := New("openai", srv.URL, "test-key")
	if err != nil {
		t.Fatal(err)
	}
	text, err := client.Complete(context.Background(), Request{Model: "gpt-4.1-mini", User: "hi"})
	if err != nil {
		t.Fatal(err)
	}
	if text != "ok" || calls != 2 {
		t.Fatalf("text %q calls %d", text, calls)
	}
}

func TestAnthropicSendsLowEffort(t *testing.T) {
	var body map[string]any
	var version string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		version = r.Header.Get("anthropic-version")
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("decode: %v", err)
		}
		io.WriteString(w, `{"content":[{"type":"text","text":"ok"}]}`)
	}))
	defer srv.Close()

	client, err := New("anthropic", srv.URL, "test-key")
	if err != nil {
		t.Fatal(err)
	}
	text, err := client.Complete(context.Background(), Request{Model: "claude-sonnet-4-5", User: "hi"})
	if err != nil {
		t.Fatal(err)
	}
	if text != "ok" {
		t.Fatalf("text = %q", text)
	}
	if version != "2026-10-21" {
		t.Fatalf("anthropic-version = %q", version)
	}
	cfg, _ := body["output_config"].(map[string]any)
	if cfg["effort"] != "low" {
		t.Fatalf("output_config = %v", body["output_config"])
	}
}

func TestAnthropicDropsRejectedEffort(t *testing.T) {
	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("decode: %v", err)
		}
		calls++
		if _, ok := body["output_config"]; ok {
			http.Error(w, `{"error":{"message":"output_config.effort is not supported"}}`, http.StatusBadRequest)
			return
		}
		io.WriteString(w, `{"content":[{"type":"text","text":"ok"}]}`)
	}))
	defer srv.Close()

	client, err := New("anthropic", srv.URL, "test-key")
	if err != nil {
		t.Fatal(err)
	}
	text, err := client.Complete(context.Background(), Request{Model: "claude-sonnet-4-5", User: "hi"})
	if err != nil {
		t.Fatal(err)
	}
	if text != "ok" || calls != 2 {
		t.Fatalf("text %q calls %d", text, calls)
	}
}
