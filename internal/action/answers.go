package action

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/shipquiz/shipquiz/internal/quiz"
)

// Marker is the machine-readable payload hidden in the quiz comment.
type Marker struct {
	V          int            `json:"v"`
	HeadSHA    string         `json:"head_sha"`
	CheckRunID int64          `json:"check_run_id"`
	Quiz       quiz.Generated `json:"quiz"`
}

// ParseAnswers reads q1: / 1. lines from a PR reply.
func ParseAnswers(body string, questions []quiz.Question) []quiz.Answer {
	byID := map[string]string{}
	var current string
	var buf strings.Builder
	flush := func() {
		if current == "" {
			return
		}
		byID[current] = strings.TrimSpace(buf.String())
		buf.Reset()
	}
	for _, line := range strings.Split(body, "\n") {
		if id, rest, ok := matchAnswerLine(line, questions); ok {
			flush()
			current = id
			buf.WriteString(rest)
			continue
		}
		if current != "" {
			if buf.Len() > 0 {
				buf.WriteByte('\n')
			}
			buf.WriteString(line)
		}
	}
	flush()
	answers := make([]quiz.Answer, 0, len(questions))
	for _, question := range questions {
		answers = append(answers, quiz.Answer{ID: question.ID, Text: strings.TrimSpace(byID[question.ID])})
	}
	return answers
}

// HasAnyAnswer reports whether the reply included at least one non-empty answer.
func HasAnyAnswer(answers []quiz.Answer) bool {
	for _, answer := range answers {
		if strings.TrimSpace(answer.Text) != "" {
			return true
		}
	}
	return false
}

func matchAnswerLine(line string, questions []quiz.Question) (string, string, bool) {
	trimmed := strings.TrimSpace(line)
	lower := strings.ToLower(trimmed)
	for i, question := range questions {
		id := strings.ToLower(question.ID)
		prefixes := []string{
			id + ":",
			id + ")",
			fmt.Sprintf("%d.", i+1),
			fmt.Sprintf("%d)", i+1),
			fmt.Sprintf("%d:", i+1),
		}
		for _, prefix := range prefixes {
			if strings.HasPrefix(lower, prefix) {
				return question.ID, strings.TrimSpace(trimmed[len(prefix):]), true
			}
		}
	}
	return "", "", false
}

func encodeMarker(marker Marker) (string, error) {
	raw, err := json.Marshal(marker)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(raw), nil
}

func decodeMarker(body string) (Marker, bool) {
	const prefix = "<!-- shipquiz "
	start := strings.Index(body, prefix)
	if start < 0 {
		return Marker{}, false
	}
	rest := body[start+len(prefix):]
	end := strings.Index(rest, "-->")
	if end < 0 {
		return Marker{}, false
	}
	raw, err := base64.StdEncoding.DecodeString(strings.TrimSpace(rest[:end]))
	if err != nil {
		return Marker{}, false
	}
	var marker Marker
	if err := json.Unmarshal(raw, &marker); err != nil {
		return Marker{}, false
	}
	return marker, true
}

func renderQuiz(marker Marker) (string, error) {
	encoded, err := encodeMarker(marker)
	if err != nil {
		return "", err
	}
	var b strings.Builder
	fmt.Fprintf(&b, "<!-- shipquiz %s -->\n\n", encoded)
	fmt.Fprintf(&b, "## shipquiz\n\n**%s** — %s\n\n", blank(marker.Quiz.Size, "change"), marker.Quiz.Summary)
	b.WriteString("Reply in this thread. Start each answer with its id (`q1:`, `q2:`, …).\n\n")
	for _, question := range marker.Quiz.Questions {
		fmt.Fprintf(&b, "### %s\n\n%s\n\n", question.ID, question.Question)
		if question.WhyItMatters != "" {
			fmt.Fprintf(&b, "_Why it matters:_ %s\n\n", question.WhyItMatters)
		}
	}
	return b.String(), nil
}

func renderResult(ev quiz.Evaluation) string {
	var b strings.Builder
	fmt.Fprintf(&b, "## shipquiz · %s\n\nScore **%.1f / 2**\n\n%s\n", ev.Overall, ev.Score, ev.Feedback)
	for _, item := range ev.PerQuestion {
		fmt.Fprintf(&b, "\n- **%s** %d — %s", item.ID, item.Score, item.Comment)
	}
	b.WriteString("\n")
	return b.String()
}

func blank(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}
