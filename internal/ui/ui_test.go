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

func TestCtrlGShowsAnswerWithoutSubmitting(t *testing.T) {
	var out bytes.Buffer
	in := bufio.NewReader(strings.NewReader("hel\x07lo\n\n"))
	got, err := editAnswer(in, nil, &out, "The cache owns invalidation.", false)
	if err != nil {
		t.Fatal(err)
	}
	if got != "hello" {
		t.Fatalf("answer = %q", got)
	}
	if !strings.Contains(out.String(), "The cache owns invalidation.") {
		t.Fatalf("output:\n%s", out.String())
	}
}

func TestBlankLineSubmits(t *testing.T) {
	var out bytes.Buffer
	in := bufio.NewReader(strings.NewReader("one\ntwo\n\n"))
	got, err := editAnswer(in, nil, &out, "", false)
	if err != nil {
		t.Fatal(err)
	}
	if got != "one\ntwo" {
		t.Fatalf("answer = %q", got)
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
