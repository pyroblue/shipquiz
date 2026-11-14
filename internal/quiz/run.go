package quiz

import (
	"context"
	"fmt"
	"time"

	"github.com/shipquiz/shipquiz/internal/llm"
)

// Generate asks the model for questions and normalizes the result.
func Generate(ctx context.Context, client llm.Client, model string, diff Diff, style string, minQ, maxQ int) (Generated, error) {
	system, user := GenerationPrompt(diff, style, minQ, maxQ)
	max := 900
	if SubstantiveLines(diff.Text) > 80 {
		max = 1600
	}
	raw, err := client.Complete(ctx, llm.Request{
		Model:     model,
		System:    system,
		User:      user,
		MaxTokens: max,
	})
	if err != nil {
		return Generated{}, err
	}
	g, err := ParseGenerated(raw)
	if err != nil {
		return Generated{}, err
	}
	g = Normalize(g, minQ, maxQ)
	if len(g.Questions) == 0 {
		return Generated{}, fmt.Errorf("the model returned no questions")
	}
	return g, nil
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

// WithTimeout returns a context limited to seconds, or 20 when seconds is unset.
func WithTimeout(parent context.Context, seconds int) (context.Context, context.CancelFunc) {
	if seconds <= 0 {
		seconds = 20
	}
	return context.WithTimeout(parent, time.Duration(seconds)*time.Second)
}
