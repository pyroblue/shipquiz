package quiz

import (
	"encoding/json"
	"fmt"
	"strings"
)

const generationSystem = "You output only JSON. No markdown fences, no preamble."

const generationTemplate = `You are an expert software architect. Generate questions that test whether the developer understands the important architectural decisions in this diff.

First classify the change:
- Trivial: mechanical, tests/docs/config/formatting, or < ~15 substantive lines
- Tiny: small localized, low architectural impact
- Small: single-component logic
- Medium: multiple components or new abstractions
- Large: new subsystem, redesign, or cross-cutting

Question limits:
- Trivial → 1
- Tiny → up to 2
- Small → up to 3
- Medium → usually up to 3; 4 only for four distinct decisions
- Large → usually up to 3; 4 or 5 only for that many distinct decisions

These are maximums, not targets. Ask one question per distinct, meaningful
architectural decision. Return fewer questions when the diff has fewer decisions.
Diff size alone never justifies more questions.

Select questions:
1. Privately identify the architectural decisions in the diff.
2. Rank them by user impact, security, reliability, and long-term maintenance.
3. Group related edits into one decision.
4. Discard weak candidates. Do not output this analysis.

Importance gate:
- A question must explain user-visible behavior, a system boundary, a security or reliability guarantee, or a long-lived design trade-off.
- Ask what a maintainer should understand months later, not how this patch happens to implement it today.
- If a candidate does not pass this gate, omit it.

Clarity gate:
- Write for a developer who understands the product but has not memorized this patch.
- Name the subject in plain language and ask one thing.
- Keep each question to one sentence and roughly 18 words or fewer.
- Avoid “why X but Y?” constructions.
- Avoid implementation verbs such as strip, marshal, render, retry, or pass.
- Avoid package, function, type, field, parameter, provider, and keyboard-shortcut names unless essential to the architecture.
- Reject questions answered by merely restating the diff or a product requirement.
- Reject questions about quiz mechanics unless they define important user behavior.

Choices:
- Give exactly three short, mutually exclusive choices.
- One choice is correct. The other two are plausible misconceptions.
- Avoid trick wording, “all of the above,” and obviously silly distractors.
- Keep choices similar in length and vary the position of the correct choice.
- explanation teaches why the correct choice is right in one short sentence.

Rules:
- Produce between %d and %d questions. Never exceed %d. Never return zero questions.
- Question style: %s.
- If the style is failure-modes, ask about an important failure and the guarantee that contains it.
- If the style is data-flow, ask about an important system boundary that data crosses.
- If the style is default, follow the rules above as written.
- Output only JSON:

{
  "size": "trivial|tiny|small|medium|large",
  "summary": "one sentence",
  "questions": [
    {
      "id": "q1",
      "question": "...",
      "choices": ["...", "...", "..."],
      "correct_choice": 2,
      "explanation": "..."
    }
  ]
}

Diff:
%s

Context:
%s
`

const evaluationSystem = "You output only JSON. No markdown fences, no preamble. Score only from the diff and the answers."

const evaluationTemplate = `Score the developer’s answers. Be strict but fair.

Judge architectural understanding. A short answer that names the right piece and its role is a 2.
Questions may be multiple choice. Resolve a numbered answer against its choices,
then use the diff to decide whether that choice is correct.

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

const interpretSystem = "You output only JSON. No markdown fences, no preamble."

const interpretTemplate = `A reviewer replied to a shipquiz comment. Decide whether the reply tries to answer the quiz, and copy out each answer.

The reply may be prose, labeled, or a list of selected choice numbers in
question order, such as "2, 1, 3". Labels are not required.

attempted is false when the comment is not answering the quiz, such as "LGTM", a question back to the author, or unrelated discussion.
attempted is true when the reply tries to answer one or more questions, even if some answers are missing or wrong.

When attempted is true, return one entry per question id. text is the selected
choice number or the part of the reply that answers that question, copied from
the reply. Use an empty string when that question was not answered. Do not
invent an answer.

Output only JSON:

{
  "attempted": true,
  "answers": [
    {"id": "q1", "text": "..."}
  ]
}

Questions:
%s

Reply:
%s
`

// InterpretPrompt builds the messages that map a freeform comment onto questions.
func InterpretPrompt(questions []Question, comment string) (system, user string) {
	q, _ := json.MarshalIndent(questions, "", "  ")
	user = fmt.Sprintf(interpretTemplate, string(q), comment)
	return interpretSystem, user
}
