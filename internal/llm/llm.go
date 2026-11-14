// Package llm calls the user's own model provider. Keys are passed in;
// this package never stores them.
package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/shipquiz/shipquiz/internal/version"
)

// Request is one completion.
type Request struct {
	Model     string
	System    string
	User      string
	MaxTokens int
}

// Client generates text and lists models for one provider.
type Client interface {
	Complete(ctx context.Context, req Request) (string, error)
	ListModels(ctx context.Context) ([]string, error)
}

// New builds a client for provider. baseURL is optional.
// anthropic expects an origin (https://api.anthropic.com).
// openai, gemini, and xai expect an API root that already includes /v1 when required.
func New(provider, baseURL, apiKey string) (Client, error) {
	if strings.TrimSpace(apiKey) == "" {
		return nil, fmt.Errorf("missing API key")
	}
	provider = strings.ToLower(strings.TrimSpace(provider))
	switch provider {
	case "", "anthropic":
		if baseURL == "" {
			baseURL = "https://api.anthropic.com"
		}
		return &anthropicClient{base: strings.TrimRight(baseURL, "/"), key: apiKey, http: newHTTP()}, nil
	case "openai":
		if baseURL == "" {
			baseURL = "https://api.openai.com/v1"
		}
		return &openAIClient{base: strings.TrimRight(baseURL, "/"), key: apiKey, http: newHTTP()}, nil
	case "gemini":
		if baseURL == "" {
			baseURL = "https://generativelanguage.googleapis.com/v1beta/openai"
		}
		return &openAIClient{base: strings.TrimRight(baseURL, "/"), key: apiKey, http: newHTTP()}, nil
	case "xai":
		if baseURL == "" {
			baseURL = "https://api.x.ai/v1"
		}
		return &openAIClient{base: strings.TrimRight(baseURL, "/"), key: apiKey, http: newHTTP()}, nil
	default:
		return nil, fmt.Errorf("unknown provider %q (use anthropic, openai, gemini, or xai)", provider)
	}
}

func newHTTP() *http.Client {
	return &http.Client{Timeout: 60 * time.Second}
}

type anthropicClient struct {
	base string
	key  string
	http *http.Client
}

func (c *anthropicClient) Complete(ctx context.Context, req Request) (string, error) {
	body := map[string]any{
		"model":      req.Model,
		"max_tokens": maxTokens(req.MaxTokens),
		"system":     req.System,
		"messages": []map[string]string{
			{"role": "user", "content": req.User},
		},
	}
	var parsed struct {
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
		Error *struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := c.post(ctx, c.base+"/v1/messages", body, &parsed); err != nil {
		return "", err
	}
	if parsed.Error != nil && parsed.Error.Message != "" {
		return "", fmt.Errorf("anthropic: %s", parsed.Error.Message)
	}
	var text strings.Builder
	for _, block := range parsed.Content {
		if block.Type == "text" || block.Text != "" {
			text.WriteString(block.Text)
		}
	}
	if text.Len() == 0 {
		return "", fmt.Errorf("anthropic returned an empty response")
	}
	return text.String(), nil
}

func (c *anthropicClient) ListModels(ctx context.Context) ([]string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.base+"/v1/models", nil)
	if err != nil {
		return nil, err
	}
	c.auth(req)
	return decodeModelIDs(c.http.Do(req))
}

func (c *anthropicClient) post(ctx context.Context, url string, body any, out any) error {
	raw, err := json.Marshal(body)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(raw))
	if err != nil {
		return err
	}
	c.auth(req)
	req.Header.Set("content-type", "application/json")
	res, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	data, err := io.ReadAll(io.LimitReader(res.Body, 1<<20))
	if err != nil {
		return err
	}
	if res.StatusCode >= 300 {
		return fmt.Errorf("anthropic returned %d: %s", res.StatusCode, trimBody(data))
	}
	if out == nil {
		return nil
	}
	return json.Unmarshal(data, out)
}

func (c *anthropicClient) auth(req *http.Request) {
	req.Header.Set("x-api-key", c.key)
	req.Header.Set("anthropic-version", "2023-06-01")
	req.Header.Set("user-agent", "shipquiz/"+version.Current)
}

type openAIClient struct {
	base string
	key  string
	http *http.Client
}

func (c *openAIClient) Complete(ctx context.Context, req Request) (string, error) {
	text, err := c.complete(ctx, req, true)
	if err == nil {
		return text, nil
	}
	if strings.Contains(err.Error(), "max_completion_tokens") {
		return c.complete(ctx, req, false)
	}
	return "", err
}

func (c *openAIClient) complete(ctx context.Context, req Request, completionTokens bool) (string, error) {
	body := map[string]any{
		"model": req.Model,
		"messages": []map[string]string{
			{"role": "system", "content": req.System},
			{"role": "user", "content": req.User},
		},
	}
	if completionTokens {
		body["max_completion_tokens"] = maxTokens(req.MaxTokens)
	} else {
		body["max_tokens"] = maxTokens(req.MaxTokens)
	}
	raw, err := json.Marshal(body)
	if err != nil {
		return "", err
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.base+"/chat/completions", bytes.NewReader(raw))
	if err != nil {
		return "", err
	}
	c.auth(httpReq)
	httpReq.Header.Set("content-type", "application/json")
	res, err := c.http.Do(httpReq)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	data, err := io.ReadAll(io.LimitReader(res.Body, 1<<20))
	if err != nil {
		return "", err
	}
	if res.StatusCode >= 300 {
		return "", fmt.Errorf("openai returned %d: %s", res.StatusCode, trimBody(data))
	}
	var parsed struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Error *struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(data, &parsed); err != nil {
		return "", err
	}
	if parsed.Error != nil && parsed.Error.Message != "" {
		return "", fmt.Errorf("openai: %s", parsed.Error.Message)
	}
	if len(parsed.Choices) == 0 || strings.TrimSpace(parsed.Choices[0].Message.Content) == "" {
		return "", fmt.Errorf("openai returned an empty response")
	}
	return parsed.Choices[0].Message.Content, nil
}

func (c *openAIClient) ListModels(ctx context.Context) ([]string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.base+"/models", nil)
	if err != nil {
		return nil, err
	}
	c.auth(req)
	return decodeModelIDs(c.http.Do(req))
}

func (c *openAIClient) auth(req *http.Request) {
	req.Header.Set("authorization", "Bearer "+c.key)
	req.Header.Set("user-agent", "shipquiz/"+version.Current)
}

func decodeModelIDs(res *http.Response, err error) ([]string, error) {
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	data, err := io.ReadAll(io.LimitReader(res.Body, 1<<20))
	if err != nil {
		return nil, err
	}
	if res.StatusCode >= 300 {
		return nil, fmt.Errorf("model list returned %d: %s", res.StatusCode, trimBody(data))
	}
	var parsed struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(data, &parsed); err != nil {
		return nil, err
	}
	seen := map[string]struct{}{}
	var ids []string
	for _, item := range parsed.Data {
		id := strings.TrimSpace(item.ID)
		if id == "" || skipModel(id) {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		ids = append(ids, id)
	}
	sort.Strings(ids)
	if len(ids) == 0 {
		return nil, fmt.Errorf("provider returned no chat models")
	}
	return ids, nil
}

func skipModel(id string) bool {
	lower := strings.ToLower(id)
	for _, part := range []string{"embed", "tts", "whisper", "dall-e", "moderation", "realtime", "image", "imagine", "voice"} {
		if strings.Contains(lower, part) {
			return true
		}
	}
	return false
}

func maxTokens(n int) int {
	if n <= 0 {
		return 800
	}
	return n
}

func trimBody(data []byte) string {
	msg := strings.TrimSpace(string(data))
	if len(msg) > 400 {
		return msg[:400]
	}
	return msg
}
