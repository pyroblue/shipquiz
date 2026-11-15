package ui

import (
	"bufio"
	"bytes"
	"strings"
	"testing"

	"github.com/shipquiz/shipquiz/internal/quiz"
)

func TestWrapKeepsLinesShort(t *testing.T) {
	text := "Which component owns cache invalidation when a write crosses the API boundary in this change?"
	lines := wrap(text, questionWidth)
	if len(lines) < 2 {
		t.Fatalf("expected a wrap, got %#v", lines)
	}
	for _, line := range lines {
		if len(line) > questionWidth {
			t.Fatalf("line %d: %q", len(line), line)
		}
	}
	if strings.Join(lines, " ") != text {
		t.Fatalf("lost words: %#v", lines)
	}
}

func TestQuestionLabelKeepsOriginalNumberOnRetry(t *testing.T) {
	if got := questionLabel("q4", 1); got != "4" {
		t.Fatalf("label = %q", got)
	}
}

func TestCollectChoicesScoresSingleKeySelection(t *testing.T) {
	var out bytes.Buffer
	in := bufio.NewReader(strings.NewReader("2\n"))
	questions := []quiz.Question{{
		ID:            "q1",
		Question:      "Who owns the boundary?",
		Choices:       []string{"The CLI", "The API", "The cache"},
		CorrectChoice: 2,
		Explanation:   "The API owns the boundary.",
	}}
	answers, err := CollectChoices(in, nil, &out, questions)
	if err != nil {
		t.Fatal(err)
	}
	if len(answers) != 1 || answers[0].Choice != 2 || answers[0].Text != "The API" {
		t.Fatalf("answers = %#v", answers)
	}
	if !strings.Contains(out.String(), "correct") {
		t.Fatalf("output:\n%s", out.String())
	}
}

func TestQuestionMarkRevealsAndCountsIncorrect(t *testing.T) {
	var out bytes.Buffer
	in := bufio.NewReader(strings.NewReader("?\n"))
	questions := []quiz.Question{{
		ID:            "q1",
		Question:      "Who owns the boundary?",
		Choices:       []string{"The CLI", "The API", "The cache"},
		CorrectChoice: 2,
		Explanation:   "The API owns the boundary.",
	}}
	answers, err := CollectChoices(in, nil, &out, questions)
	if err != nil {
		t.Fatal(err)
	}
	if answers[0].Choice != 0 {
		t.Fatalf("answers = %#v", answers)
	}
	if !strings.Contains(out.String(), "2 · The API") {
		t.Fatalf("output:\n%s", out.String())
	}
}

func TestChooseMissActionReadsTheMenu(t *testing.T) {
	var out bytes.Buffer
	in := bufio.NewReader(strings.NewReader("1\n"))
	action, err := ChooseMissAction(in, nil, &out, true, true)
	if err != nil {
		t.Fatal(err)
	}
	if action != MissPush {
		t.Fatalf("action = %v", action)
	}
	text := out.String()
	for _, want := range []string{"push anyway", "retry missed questions", "restart this quiz", "generate new questions"} {
		if !strings.Contains(text, want) {
			t.Fatalf("menu missing %q:\n%s", want, text)
		}
	}

	in = bufio.NewReader(strings.NewReader("2\n"))
	action, err = ChooseMissAction(in, nil, &bytes.Buffer{}, false, true)
	if err == nil || action != MissStop {
		t.Fatalf("hidden retry action=%v err=%v", action, err)
	}
}

func TestPassingResultCelebratesWithoutAnimatingAPipe(t *testing.T) {
	var out bytes.Buffer
	PrintResult(&out, quiz.Evaluation{Overall: "pass", Score: 2})

	got := out.String()
	if !strings.Contains(got, "PASS") || !strings.Contains(got, "cleared to ship") {
		t.Fatalf("output:\n%s", got)
	}
	if strings.Contains(got, "\033") {
		t.Fatalf("animation leaked into a pipe: %q", got)
	}
}

func TestQuizResultLinesStayReadable(t *testing.T) {
	var out bytes.Buffer
	PrintResult(&out, quiz.Evaluation{
		Overall: "fail",
		Score:   1,
		Feedback: "The answers identify the broad goal but miss the reliability " +
			"guarantee that controls slow model requests across interactive and hook runs.",
		PerQuestion: []quiz.PerQuestion{{
			ID:    "q1",
			Score: 1,
			Comment: "The answer names the model boundary but does not explain " +
				"which user-facing behavior remains reliable when that boundary is slow.",
		}},
	})

	for _, line := range strings.Split(strings.TrimSpace(out.String()), "\n") {
		if len([]rune(line)) > questionWidth {
			t.Fatalf("line is %d columns:\n%s", len([]rune(line)), line)
		}
	}
}
