package quiz

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/shipquiz/shipquiz/internal/llm"
)

// ModelOutput is a generation failure that keeps the model's response.
type ModelOutput struct {
	Reason string
	Raw    string
}

// Error prints the reason and the response the model sent.
func (e *ModelOutput) Error() string {
	body := strings.TrimSpace(e.Raw)
	if body == "" {
		return e.Reason
	}
	return e.Reason + "\n\n" + body
}

// Generate asks the model for questions and normalizes the result.
func Generate(ctx context.Context, client llm.Client, model string, diff Diff, style string, minQ, maxQ int) (Generated, error) {
	system, user := GenerationPrompt(diff, style, minQ, maxQ)
	max := 900
	if SubstantiveLines(diff.Text) > 80 {
		max = 1600
	}
	raw, err := client.Complete(ctx, llm.Request{
		Model:      model,
		System:     system,
		User:       user,
		MaxTokens:  max,
		JSONObject: true,
	})
	if err != nil {
		return Generated{}, err
	}
	g, err := acceptGenerated(raw, minQ, maxQ)
	if err == nil {
		return g, nil
	}
	var out *ModelOutput
	if !errors.As(err, &out) {
		return Generated{}, err
	}
	repaired, retryErr := client.Complete(ctx, llm.Request{
		Model:      model,
		System:     generationSystem,
		User:       repairQuizPrompt(out),
		MaxTokens:  max,
		JSONObject: true,
	})
	if retryErr != nil {
		return Generated{}, out
	}
	g, err = acceptGenerated(repaired, minQ, maxQ)
	if err != nil {
		return Generated{}, err
	}
	return g, nil
}

// acceptGenerated parses, normalizes, and validates a quiz response.
func acceptGenerated(raw string, minQ, maxQ int) (Generated, error) {
	g, err := ParseGenerated(raw)
	if err != nil {
		return Generated{}, &ModelOutput{Reason: err.Error(), Raw: raw}
	}
	g = Normalize(g, minQ, maxQ)
	if len(g.Questions) == 0 {
		return Generated{}, &ModelOutput{Reason: "the model returned no questions", Raw: raw}
	}
	if issue := ChoiceIssue(g); issue != "" {
		return Generated{}, &ModelOutput{
			Reason: "the model returned an invalid multiple-choice quiz\n" + issue,
			Raw:    raw,
		}
	}
	RotateChoices(&g)
	return g, nil
}

// repairQuizPrompt asks the model to fix an invalid quiz response.
func repairQuizPrompt(out *ModelOutput) string {
	return fmt.Sprintf(`The quiz JSON was invalid:
%s

Return corrected JSON only. Each question needs exactly three choices.
correct_choice must be 1, 2, or 3 (1-based). Never use 0.

Invalid response:
%s
`, out.Reason, strings.TrimSpace(out.Raw))
}

// Grade scores answers and applies the local pass rule.
func Grade(ctx context.Context, client llm.Client, model string, diff Diff, questions []Question, answers []Answer, passing float64) (Evaluation, error) {
	system, user := EvaluationPrompt(diff, questions, answers, passing)
	raw, err := client.Complete(ctx, llm.Request{
		Model:     model,
		System:    system,
		User:      user,
		MaxTokens: 900,
	})
	if err != nil {
		return Evaluation{}, err
	}
	ev, err := ParseEvaluation(raw)
	if err != nil {
		return Evaluation{}, err
	}
	return Enforce(ev, questions, passing), nil
}

// Interpret asks the model whether a freeform comment answers the quiz.
func Interpret(ctx context.Context, client llm.Client, model string, questions []Question, comment string) (Reply, error) {
	if strings.TrimSpace(comment) == "" {
		return Reply{}, nil
	}
	system, user := InterpretPrompt(questions, comment)
	raw, err := client.Complete(ctx, llm.Request{
		Model:     model,
		System:    system,
		User:      user,
		MaxTokens: 900,
	})
	if err != nil {
		return Reply{}, err
	}
	reply, err := ParseReply(raw)
	if err != nil {
		return Reply{}, err
	}
	return NormalizeReply(reply, questions), nil
}

// WithTimeout returns a context limited to seconds, or 60 when seconds is unset.
func WithTimeout(parent context.Context, seconds int) (context.Context, context.CancelFunc) {
	if seconds <= 0 {
		seconds = 60
	}
	return context.WithTimeout(parent, time.Duration(seconds)*time.Second)
}
