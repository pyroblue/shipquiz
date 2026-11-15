package quiz

import (
	"crypto/sha256"
	"fmt"
	"strings"
)

// Question is one item shown to the developer.
type Question struct {
	ID            string   `json:"id"`
	Question      string   `json:"question"`
	WhyItMatters  string   `json:"why_it_matters,omitempty"`
	Reference     string   `json:"reference,omitempty"`
	Choices       []string `json:"choices,omitempty"`
	CorrectChoice int      `json:"correct_choice,omitempty"`
	Explanation   string   `json:"explanation,omitempty"`
}

// Generated is the model’s quiz for a diff.
type Generated struct {
	Size      string     `json:"size"`
	Summary   string     `json:"summary"`
	Questions []Question `json:"questions"`
}

// Answer is the developer’s reply to one question.
type Answer struct {
	ID     string `json:"id"`
	Text   string `json:"text,omitempty"`
	Choice int    `json:"choice,omitempty"`
}

// Reply is a model's reading of a freeform pull-request comment.
type Reply struct {
	Attempted bool     `json:"attempted"`
	Answers   []Answer `json:"answers"`
}

// PerQuestion is the grade for one answer.
type PerQuestion struct {
	ID      string `json:"id"`
	Score   int    `json:"score"`
	Comment string `json:"comment"`
}

// Evaluation is the graded quiz after the local pass rule is applied.
type Evaluation struct {
	Overall     string        `json:"overall"`
	Score       float64       `json:"score"`
	Feedback    string        `json:"feedback"`
	PerQuestion []PerQuestion `json:"per_question"`
}

// Passed reports whether the quiz cleared the bar.
func (e Evaluation) Passed() bool {
	return e.Overall == "pass"
}

// Enforce applies the v1 pass rule: average >= passing and no zero scores.
// The model’s overall field is replaced by this result.
func Enforce(ev Evaluation, questions []Question, passing float64) Evaluation {
	byID := map[string]PerQuestion{}
	for _, item := range ev.PerQuestion {
		item.Score = clampScore(item.Score)
		byID[item.ID] = item
	}
	ev.PerQuestion = ev.PerQuestion[:0]
	sum := 0
	hasZero := false
	for i, q := range questions {
		item, ok := byID[q.ID]
		if !ok {
			item = PerQuestion{ID: q.ID, Score: 0, Comment: "No score returned for this answer."}
		}
		if item.Score == 0 {
			hasZero = true
		}
		sum += item.Score
		ev.PerQuestion = append(ev.PerQuestion, item)
		_ = i
	}
	if len(questions) == 0 {
		ev.Score = 0
		ev.Overall = "fail"
		return ev
	}
	ev.Score = float64(sum) / float64(len(questions))
	if ev.Score+1e-9 >= passing && !hasZero {
		ev.Overall = "pass"
	} else {
		ev.Overall = "fail"
	}
	return ev
}

// GradeChoices scores selected choices locally against the generated answer key.
func GradeChoices(questions []Question, answers []Answer) Evaluation {
	byID := make(map[string]Answer, len(answers))
	for _, answer := range answers {
		byID[answer.ID] = answer
	}
	ev := Evaluation{
		Overall:  "pass",
		Feedback: "Every choice is correct.",
	}
	for _, question := range questions {
		answer := byID[question.ID]
		item := PerQuestion{ID: question.ID}
		if answer.Choice == question.CorrectChoice && question.CorrectChoice > 0 {
			item.Score = 2
			item.Comment = "Correct."
		} else {
			item.Score = 0
			item.Comment = question.Explanation
			if item.Comment == "" {
				item.Comment = "Choose the answer that best explains the architectural decision."
			}
			ev.Overall = "fail"
		}
		ev.PerQuestion = append(ev.PerQuestion, item)
	}
	if ev.Overall == "fail" {
		ev.Feedback = "Review the explanation, then retake the missed questions."
	}
	return Enforce(ev, questions, 2)
}

// MultipleChoiceReady reports whether every question has a usable answer key.
func MultipleChoiceReady(g Generated) bool {
	return len(g.Questions) > 0 && ChoiceIssue(g) == ""
}

// ChoiceIssue explains why a parsed quiz cannot be offered as multiple choice.
func ChoiceIssue(g Generated) string {
	var lines []string
	for _, question := range g.Questions {
		id := strings.TrimSpace(question.ID)
		if id == "" {
			id = "question"
		}
		if len(question.Choices) != 3 {
			lines = append(lines, fmt.Sprintf("%s: expected 3 choices, got %d", id, len(question.Choices)))
		}
		if question.CorrectChoice < 1 || question.CorrectChoice > 3 {
			lines = append(lines, fmt.Sprintf("%s: correct_choice must be 1, 2, or 3 (got %d)", id, question.CorrectChoice))
		}
		for i, choice := range question.Choices {
			if strings.TrimSpace(choice) == "" {
				lines = append(lines, fmt.Sprintf("%s: choice %d is empty", id, i+1))
			}
		}
	}
	return strings.Join(lines, "\n")
}

// RotateChoices distributes correct answers across positions without changing the key.
func RotateChoices(g *Generated) {
	for i := range g.Questions {
		question := &g.Questions[i]
		if len(question.Choices) != 3 || question.CorrectChoice < 1 || question.CorrectChoice > 3 {
			continue
		}
		sum := sha256.Sum256([]byte(question.Question))
		offset := int(sum[0]) % len(question.Choices)
		if offset == 0 {
			continue
		}
		choices := append([]string(nil), question.Choices...)
		for j, choice := range choices {
			question.Choices[(j+offset)%len(choices)] = choice
		}
		question.CorrectChoice = (question.CorrectChoice-1+offset)%len(choices) + 1
	}
}

func clampScore(n int) int {
	if n < 0 {
		return 0
	}
	if n > 2 {
		return 2
	}
	return n
}

// Normalize fills missing ids and trims the question list to [minQ, maxQ].
func Normalize(g Generated, minQ, maxQ int) Generated {
	if maxQ > 5 {
		maxQ = 5
	}
	if maxQ < 1 {
		maxQ = 1
	}
	if minQ < 1 {
		minQ = 1
	}
	if minQ > maxQ {
		minQ = maxQ
	}
	var questions []Question
	for _, q := range g.Questions {
		q.Question = strings.TrimSpace(q.Question)
		q.WhyItMatters = strings.TrimSpace(q.WhyItMatters)
		q.Reference = strings.TrimSpace(q.Reference)
		q.Explanation = strings.TrimSpace(q.Explanation)
		for i := range q.Choices {
			q.Choices[i] = strings.TrimSpace(q.Choices[i])
		}
		if q.Question == "" {
			continue
		}
		questions = append(questions, q)
	}
	if len(questions) > maxQ {
		questions = questions[:maxQ]
	}
	for i := range questions {
		if strings.TrimSpace(questions[i].ID) == "" {
			questions[i].ID = "q" + itoa(i+1)
		}
	}
	g.Questions = questions
	g.Size = strings.TrimSpace(g.Size)
	g.Summary = strings.TrimSpace(g.Summary)
	return g
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var digits [4]byte
	i := len(digits)
	for n > 0 {
		i--
		digits[i] = byte('0' + n%10)
		n /= 10
	}
	return string(digits[i:])
}
