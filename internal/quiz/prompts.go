package quiz

import (
	"encoding/json"
	"fmt"
	"strings"
)

const generationSystem = "You output only JSON. No markdown fences, no preamble."

const generationTemplate = `You are an expert software architect. Generate questions that test whether the developer deeply understands the architectural implications of this diff.

First classify the change:
- Trivial: mechanical, tests/docs/config/formatting, or < ~15 substantive lines
- Tiny: small localized, low architectural impact
- Small: single-component logic
- Medium: multiple components or new abstractions
- Large: new subsystem, redesign, or cross-cutting

Generate this many questions:
- Trivial → 1
- Tiny → 2
- Small → 3
- Medium → 4 or 5
- Large → 6 or 7

Rules:
- Every question must be specific to this diff.
- Focus on data flow, invariants, failure modes, boundaries, and design trade-offs.
- Answerable in 1–3 sentences.
- No syntax/style questions.
- Prefer “what happens if…”, “why here instead of there”, “walk through…”.
- Produce between %d and %d questions. Never exceed %d. Never return zero questions.
- Question style: %s.
- If the style is failure-modes, every question asks about a failure mode or a broken invariant.
- If the style is data-flow, every question asks the developer to walk data across a boundary in this diff.
- If the style is default, follow the rules above as written.
- Output only JSON:

{
  "size": "trivial|tiny|small|medium|large",
  "summary": "one sentence",
  "questions": [
    {"id": "q1", "question": "...", "why_it_matters": "..."}
  ]
}

Diff:
%s

Context:
%s
`

const evaluationSystem = "You output only JSON. No markdown fences, no preamble. Score only from the diff and the answers."

const evaluationTemplate = `Score the developer’s answers. Be strict but fair.

Per question: 0 (wrong/missing), 1 (partial), 2 (clear understanding).

Overall "pass" only if average >= %.2f AND no zeros. Otherwise "fail".

Output only JSON:

{
  "overall": "pass" | "fail",
  "score": 1.8,
  "feedback": "2-4 sentence critique",
  "per_question": [
    {"id": "q1", "score": 2, "comment": "..."}
  ]
}

Questions:
%s

Answers:
%s

Diff (reference):
%s
`

// GenerationPrompt builds the question-generation messages.
func GenerationPrompt(diff Diff, style string, minQ, maxQ int) (system, user string) {
	if style == "" {
		style = "default"
	}
	context := fmt.Sprintf("Changed files (%d):\n%s\n\nSubstantive added/removed lines: %d\n\n%s",
		len(diff.Files),
		strings.Join(diff.Files, "\n"),
		SubstantiveLines(diff.Text),
		strings.TrimSpace(diff.Stat),
	)
	user = fmt.Sprintf(generationTemplate, minQ, maxQ, maxQ, style, diff.Text, context)
	return generationSystem, user
}

// EvaluationPrompt builds the grading messages.
func EvaluationPrompt(diff Diff, questions []Question, answers []Answer, passing float64) (system, user string) {
	q, _ := json.MarshalIndent(questions, "", "  ")
	a, _ := json.MarshalIndent(answers, "", "  ")
	user = fmt.Sprintf(evaluationTemplate, passing, string(q), string(a), diff.Text)
	return evaluationSystem, user
}
