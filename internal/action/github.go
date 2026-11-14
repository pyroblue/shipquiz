package action

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// GitHub talks to api.github.com with the Actions token.
type GitHub struct {
	Token string
	Repo  string
	HTTP  *http.Client
}

type comment struct {
	ID   int64  `json:"id"`
	Body string `json:"body"`
}

type pull struct {
	Number int `json:"number"`
	Base   struct {
		SHA string `json:"sha"`
	} `json:"base"`
	Head struct {
		SHA string `json:"sha"`
	} `json:"head"`
}

// PostComment creates an issue or pull request comment.
func (g *GitHub) PostComment(ctx context.Context, number int, body string) error {
	return g.do(ctx, http.MethodPost, fmt.Sprintf("/repos/%s/issues/%d/comments", g.Repo, number), map[string]string{"body": body}, nil)
}

// ListComments returns the first page of comments, oldest first.
func (g *GitHub) ListComments(ctx context.Context, number int) ([]comment, error) {
	var comments []comment
	err := g.do(ctx, http.MethodGet, fmt.Sprintf("/repos/%s/issues/%d/comments?per_page=100", g.Repo, number), nil, &comments)
	return comments, err
}

// GetPull loads the base and head SHAs for a pull request.
func (g *GitHub) GetPull(ctx context.Context, number int) (pull, error) {
	var pr pull
	err := g.do(ctx, http.MethodGet, fmt.Sprintf("/repos/%s/pulls/%d", g.Repo, number), nil, &pr)
	return pr, err
}

type checkRequest struct {
	Name       string `json:"name"`
	HeadSHA    string `json:"head_sha"`
	Status     string `json:"status"`
	Conclusion string `json:"conclusion,omitempty"`
	Output     struct {
		Title   string `json:"title"`
		Summary string `json:"summary"`
	} `json:"output"`
}

// CreateCheck opens a completed check run and returns its id.
func (g *GitHub) CreateCheck(ctx context.Context, headSHA, conclusion, title, summary string) (int64, error) {
	body := checkBody(headSHA, conclusion, title, summary)
	var created struct {
		ID int64 `json:"id"`
	}
	if err := g.do(ctx, http.MethodPost, fmt.Sprintf("/repos/%s/check-runs", g.Repo), body, &created); err != nil {
		if conclusion == "action_required" {
			body.Conclusion = "failure"
			if err2 := g.do(ctx, http.MethodPost, fmt.Sprintf("/repos/%s/check-runs", g.Repo), body, &created); err2 == nil {
				return created.ID, nil
			}
		}
		return 0, err
	}
	return created.ID, nil
}

// UpdateCheck sets the conclusion of an existing check run.
func (g *GitHub) UpdateCheck(ctx context.Context, id int64, headSHA, conclusion, title, summary string) error {
	body := checkBody(headSHA, conclusion, title, summary)
	return g.do(ctx, http.MethodPatch, fmt.Sprintf("/repos/%s/check-runs/%d", g.Repo, id), body, nil)
}

func checkBody(headSHA, conclusion, title, summary string) checkRequest {
	var body checkRequest
	body.Name = "shipquiz"
	body.HeadSHA = headSHA
	body.Status = "completed"
	body.Conclusion = conclusion
	body.Output.Title = title
	body.Output.Summary = summary
	return body
}

func (g *GitHub) do(ctx context.Context, method, path string, in any, out any) error {
	var payload io.Reader
	if in != nil {
		raw, err := json.Marshal(in)
		if err != nil {
			return err
		}
		payload = bytes.NewReader(raw)
	}
	req, err := http.NewRequestWithContext(ctx, method, "https://api.github.com"+path, payload)
	if err != nil {
		return err
	}
	req.Header.Set("authorization", "Bearer "+g.Token)
	req.Header.Set("accept", "application/vnd.github+json")
	req.Header.Set("x-github-api-version", "2022-11-28")
	req.Header.Set("user-agent", "shipquiz")
	if in != nil {
		req.Header.Set("content-type", "application/json")
	}
	client := g.HTTP
	if client == nil {
		client = http.DefaultClient
	}
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	data, err := io.ReadAll(io.LimitReader(res.Body, 1<<20))
	if err != nil {
		return err
	}
	if res.StatusCode >= 300 {
		msg := strings.TrimSpace(string(data))
		if len(msg) > 400 {
			msg = msg[:400]
		}
		return fmt.Errorf("github %s %s returned %d: %s", method, path, res.StatusCode, msg)
	}
	if out == nil || len(data) == 0 {
		return nil
	}
	return json.Unmarshal(data, out)
}
