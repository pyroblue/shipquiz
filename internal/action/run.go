package action

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/shipquiz/shipquiz/internal/config"
	"github.com/shipquiz/shipquiz/internal/gitx"
	"github.com/shipquiz/shipquiz/internal/llm"
	"github.com/shipquiz/shipquiz/internal/quiz"
)

type ghEvent struct {
	PullRequest *struct {
		Number int `json:"number"`
		Base   struct {
			SHA string `json:"sha"`
		} `json:"base"`
		Head struct {
			SHA string `json:"sha"`
		} `json:"head"`
	} `json:"pull_request"`
	Issue *struct {
		Number      int `json:"number"`
		PullRequest *struct {
			URL string `json:"url"`
		} `json:"pull_request"`
	} `json:"issue"`
	Comment *struct {
		Body string `json:"body"`
		User struct {
			Login string `json:"login"`
		} `json:"user"`
	} `json:"comment"`
}

// Run executes the GitHub Action entrypoint. It returns a process exit code.
func Run(ctx context.Context) int {
	eventName := os.Getenv("GITHUB_EVENT_NAME")
	eventPath := os.Getenv("GITHUB_EVENT_PATH")
	token := os.Getenv("GITHUB_TOKEN")
	repo := os.Getenv("GITHUB_REPOSITORY")
	if eventPath == "" || token == "" || repo == "" {
		fmt.Fprintln(os.Stderr, "shipquiz: action-run needs GITHUB_EVENT_PATH, GITHUB_TOKEN, and GITHUB_REPOSITORY")
		return 1
	}
	data, err := os.ReadFile(eventPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "shipquiz: read event: %v\n", err)
		return 1
	}
	var event ghEvent
	if err := json.Unmarshal(data, &event); err != nil {
		fmt.Fprintf(os.Stderr, "shipquiz: parse event: %v\n", err)
		return 1
	}
	root, _ := gitx.Root()
	cfg, err := configForAction(root)
	if err != nil {
		fmt.Fprintf(os.Stderr, "shipquiz: %v\n", err)
		return 1
	}
	apiKey := strings.TrimSpace(os.Getenv("SHIPQUIZ_LLM_API_KEY"))
	if apiKey == "" {
		fmt.Fprintln(os.Stderr, "shipquiz: set the SHIPQUIZ_LLM_API_KEY secret")
		return 1
	}
	client, err := llm.New(cfg.Provider, cfg.BaseURL, apiKey)
	if err != nil {
		fmt.Fprintf(os.Stderr, "shipquiz: %v\n", err)
		return 1
	}
	gh := &GitHub{Token: token, Repo: repo}

	switch eventName {
	case "pull_request":
		if event.PullRequest == nil {
			fmt.Fprintln(os.Stderr, "shipquiz: pull_request payload is missing")
			return 1
		}
		return postQuiz(ctx, gh, client, cfg, event.PullRequest.Number, event.PullRequest.Base.SHA, event.PullRequest.Head.SHA)
	case "issue_comment":
		if event.Issue == nil || event.Issue.PullRequest == nil || event.Comment == nil {
			fmt.Fprintln(os.Stderr, "shipquiz: this comment is not on a pull request")
			return 1
		}
		if strings.Contains(event.Comment.Body, "<!-- shipquiz ") {
			fmt.Fprintln(os.Stderr, "shipquiz: ignoring our own comment")
			return 0
		}
		return gradeReply(ctx, gh, client, cfg, event.Issue.Number, event.Comment.Body)
	default:
		fmt.Fprintf(os.Stderr, "shipquiz: unsupported event %q\n", eventName)
		return 1
	}
}

func configForAction(repoRoot string) (config.Config, error) {
	cfg, err := config.LoadForRepo(repoRoot)
	if err != nil {
		return cfg, err
	}
	if provider := strings.TrimSpace(os.Getenv("SHIPQUIZ_PROVIDER")); provider != "" {
		cfg.Provider = provider
	}
	if model := strings.TrimSpace(os.Getenv("SHIPQUIZ_MODEL")); model != "" {
		cfg.Model = model
	}
	if base := strings.TrimSpace(os.Getenv("SHIPQUIZ_BASE_URL")); base != "" {
		cfg.BaseURL = base
	}
	return cfg, nil
}

func postQuiz(ctx context.Context, gh *GitHub, client llm.Client, cfg config.Config, number int, base, head string) int {
	diff, err := ensureDiff(number, base, head)
	if err != nil {
		fmt.Fprintf(os.Stderr, "shipquiz: diff: %v\n", err)
		return 1
	}
	if diff.Empty() || quiz.ShouldSkip(diff.Files, cfg.SkipPaths) {
		if _, err := gh.CreateCheck(ctx, head, "success", "Nothing to quiz", "The diff is empty or only matches skip paths."); err != nil {
			fmt.Fprintf(os.Stderr, "shipquiz: check: %v\n", err)
			return 1
		}
		fmt.Println("shipquiz: skipped this diff")
		return 0
	}
	genCtx, cancel := quiz.WithTimeout(ctx, cfg.Quiz.TimeoutSeconds)
	defer cancel()
	generated, err := quiz.Generate(genCtx, client, cfg.Model, diff, cfg.QuestionStyle, cfg.Quiz.MinQuestions, cfg.Quiz.MaxQuestions)
	if err != nil {
		fmt.Fprintf(os.Stderr, "shipquiz: generate: %v\n", err)
		_, _ = gh.CreateCheck(ctx, head, "failure", "Could not generate the quiz", err.Error())
		return 1
	}
	checkID, err := gh.CreateCheck(ctx, head, "action_required", "Answer the quiz", generated.Summary)
	if err != nil {
		fmt.Fprintf(os.Stderr, "shipquiz: check: %v\n", err)
		checkID = 0
	}
	body, err := renderQuiz(Marker{V: 1, HeadSHA: head, CheckRunID: checkID, Quiz: generated})
	if err != nil {
		fmt.Fprintf(os.Stderr, "shipquiz: %v\n", err)
		return 1
	}
	if err := gh.PostComment(ctx, number, body); err != nil {
		fmt.Fprintf(os.Stderr, "shipquiz: comment: %v\n", err)
		return 1
	}
	fmt.Printf("shipquiz: posted %d question(s) on pull request #%d\n", len(generated.Questions), number)
	return 0
}

func gradeReply(ctx context.Context, gh *GitHub, client llm.Client, cfg config.Config, number int, reply string) int {
	comments, err := gh.ListComments(ctx, number)
	if err != nil {
		fmt.Fprintf(os.Stderr, "shipquiz: list comments: %v\n", err)
		return 1
	}
	var marker Marker
	found := false
	for _, comment := range comments {
		if decoded, ok := decodeMarker(comment.Body); ok {
			marker = decoded
			found = true
		}
	}
	if !found {
		fmt.Fprintln(os.Stderr, "shipquiz: no quiz comment found on this pull request")
		return 1
	}
	readCtx, cancelRead := quiz.WithTimeout(ctx, cfg.Quiz.TimeoutSeconds)
	read, err := quiz.Interpret(readCtx, client, cfg.Model, marker.Quiz.Questions, reply)
	cancelRead()
	if err != nil {
		fmt.Fprintf(os.Stderr, "shipquiz: read reply: %v\n", err)
		return 1
	}
	if !read.Attempted {
		fmt.Println("shipquiz: comment does not answer the quiz")
		return 0
	}
	answers := read.Answers
	pr, err := gh.GetPull(ctx, number)
	if err != nil {
		fmt.Fprintf(os.Stderr, "shipquiz: pull request: %v\n", err)
		return 1
	}
	head := marker.HeadSHA
	if pr.Head.SHA != "" {
		head = pr.Head.SHA
	}
	diff, err := ensureDiff(number, pr.Base.SHA, head)
	if err != nil {
		fmt.Fprintf(os.Stderr, "shipquiz: diff: %v\n", err)
		return 1
	}
	gradeCtx, cancel := quiz.WithTimeout(ctx, cfg.Quiz.TimeoutSeconds)
	defer cancel()
	ev, err := quiz.Grade(gradeCtx, client, cfg.Model, diff, marker.Quiz.Questions, answers, cfg.Quiz.PassingAverage)
	if err != nil {
		fmt.Fprintf(os.Stderr, "shipquiz: grade: %v\n", err)
		return 1
	}
	if err := gh.PostComment(ctx, number, renderResult(ev)); err != nil {
		fmt.Fprintf(os.Stderr, "shipquiz: result comment: %v\n", err)
		return 1
	}
	conclusion := "failure"
	title := "Quiz failed"
	if ev.Passed() {
		conclusion = "success"
		title = "Quiz passed"
	}
	summary := fmt.Sprintf("Score %.1f / 2. %s", ev.Score, ev.Feedback)
	if marker.CheckRunID > 0 {
		err = gh.UpdateCheck(ctx, marker.CheckRunID, head, conclusion, title, summary)
	} else {
		_, err = gh.CreateCheck(ctx, head, conclusion, title, summary)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "shipquiz: check: %v\n", err)
		return 1
	}
	fmt.Printf("shipquiz: %s (%.1f)\n", ev.Overall, ev.Score)
	if !ev.Passed() {
		return 1
	}
	return 0
}

func ensureDiff(number int, base, head string) (quiz.Diff, error) {
	if diff, err := quiz.RangeDiff(base, head); err == nil {
		return diff, nil
	}
	if _, err := gitx.Output("fetch", "--no-tags", "origin", fmt.Sprintf("+pull/%d/head:refs/shipquiz/pr", number)); err != nil {
		return quiz.Diff{}, err
	}
	if _, err := gitx.Output("cat-file", "-e", base+"^{commit}"); err != nil {
		_, _ = gitx.Output("fetch", "--no-tags", "origin", base)
	}
	target := head
	if _, err := gitx.Output("cat-file", "-e", head+"^{commit}"); err != nil {
		target = "refs/shipquiz/pr"
	}
	return quiz.RangeDiff(base, target)
}
