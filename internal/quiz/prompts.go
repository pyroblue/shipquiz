package quiz

import (
	"encoding/json"
	"fmt"
	"strings"
)

const generationSystem = "You output only JSON. No markdown fences, no preamble."

const generationTemplate = `You are quizzing a developer on the change in this diff. Every question must be about something this diff adds, removes, or changes. Do not ask about the rest of the application.

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

These are maximums, not targets. Ask one question per distinct change in the diff.
Return fewer questions when the diff changes fewer things.
Diff size alone never justifies more questions.

Select questions:
1. Privately list what this diff adds, removes, or changes.
2. Rank those changes by user impact, security, reliability, and how hard they are to get wrong.
3. Group related edits into one change.
4. Discard anything the diff does not change. Do not output this analysis.

Scope gate:
- A question must be answerable from this diff alone.
- Ask about the behavior, boundary, guarantee, or trade-off that this change introduces or alters.
- Do not ask about unchanged architecture, neighboring systems, or history the diff does not touch.
- If a candidate is about the application rather than this change, omit it.

Clarity gate:
- Write for a developer reading this diff.
- Name the subject in plain language and ask one thing.
- Keep each question to one sentence and roughly 18 words or fewer.
- Avoid “why X but Y?” constructions.
- Avoid line-level trivia and syntax.
- Reject questions that only quote the diff without asking what the change does.
- Reject questions about quiz mechanics unless they define important user behavior.

Choices:
- Give exactly three short, mutually exclusive choices.
- One choice is correct. The other two are plausible misconceptions.
- correct_choice is 1-based: 1, 2, or 3. Never use 0.
- Avoid trick wording, “all of the above,” and obviously silly distractors.
- Keep choices similar in length and vary the position of the correct choice.
- explanation teaches why the correct choice is right in one short sentence.

Rules:
- Produce between %d and %d questions. Never exceed %d. Never return zero questions.
- Question style: %s.
- If the style is failure-modes, ask about a failure this change introduces or closes.
- If the style is data-flow, ask about a value this change moves across a boundary in the diff.
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
