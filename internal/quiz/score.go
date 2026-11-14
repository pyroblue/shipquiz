package quiz

import "strings"

// Question is one item shown to the developer.
type Question struct {
	ID           string `json:"id"`
	Question     string `json:"question"`
	WhyItMatters string `json:"why_it_matters"`
}

// Generated is the model’s quiz for a diff.
type Generated struct {
	Size      string     `json:"size"`
	Summary   string     `json:"summary"`
	Questions []Question `json:"questions"`
}

// Answer is the developer’s reply to one question.
type Answer struct {
	ID   string `json:"id"`
	Text string `json:"text"`
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
	if maxQ > 7 {
		maxQ = 7
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
