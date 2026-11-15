package action

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"unicode"

	"github.com/shipquiz/shipquiz/internal/quiz"
)

// Marker is the machine-readable payload hidden in the quiz comment.
type Marker struct {
	V          int            `json:"v"`
	HeadSHA    string         `json:"head_sha"`
	CheckRunID int64          `json:"check_run_id"`
	Quiz       quiz.Generated `json:"quiz"`
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
	for i := range marker.Quiz.Questions {
		marker.Quiz.Questions[i].Reference = ""
		marker.Quiz.Questions[i].WhyItMatters = ""
		marker.Quiz.Questions[i].CorrectChoice = 0
		marker.Quiz.Questions[i].Explanation = ""
	}
	encoded, err := encodeMarker(marker)
	if err != nil {
		return "", err
	}
	count := len(marker.Quiz.Questions)
	countLabel := "questions"
	if count == 1 {
		countLabel = "question"
	}
	var b strings.Builder
	fmt.Fprintf(&b, "<!-- shipquiz %s -->\n\n", encoded)
	fmt.Fprintf(&b, "## shipquiz\n\n")
	fmt.Fprintf(&b, "**Change Size:** %s  \n", titleWord(blank(marker.Quiz.Size, "change")))
	fmt.Fprintf(&b, "%d %s\n\n", count, countLabel)
	fmt.Fprintf(&b, "**Diff Summary:**  \n%s\n\n", strings.TrimSpace(marker.Quiz.Summary))
	b.WriteString("Reply with the selected number for each question, in order. For example: `2, 1, 3`.\n\n")
	for i, question := range marker.Quiz.Questions {
		fmt.Fprintf(&b, "### %s\n\n%s\n\n", questionHeading(question.ID, i+1), question.Question)
		for j, choice := range question.Choices {
			fmt.Fprintf(&b, "%d. %s\n", j+1, choice)
		}
		b.WriteString("\n")
	}
	return b.String(), nil
}

func renderResult(ev quiz.Evaluation) string {
	var b strings.Builder
	b.WriteString("<!-- shipquiz result -->\n\n")
	fmt.Fprintf(&b, "## shipquiz · %s\n\nScore **%.1f / 2**\n\n%s\n", ev.Overall, ev.Score, ev.Feedback)
	for i, item := range ev.PerQuestion {
		fmt.Fprintf(&b, "\n- **%s** %d — %s", questionHeading(item.ID, i+1), item.Score, item.Comment)
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

// titleWord capitalizes the first letter of a size label.
func titleWord(word string) string {
	word = strings.TrimSpace(word)
	if word == "" {
		return word
	}
	return strings.ToUpper(word[:1]) + word[1:]
}

// questionHeading turns q1 into Q1 for the comment UI.
func questionHeading(id string, fallback int) string {
	id = strings.TrimSpace(id)
	if len(id) > 1 && (id[0] == 'q' || id[0] == 'Q') {
		return "Q" + id[1:]
	}
	if id != "" {
		return id
	}
	return fmt.Sprintf("Q%d", fallback)
}

// parseChoiceReply reads a reply like "2, 1, 3" without calling the model.
func parseChoiceReply(reply string, questions []quiz.Question) (quiz.Reply, bool) {
	fields := strings.FieldsFunc(strings.TrimSpace(reply), func(r rune) bool {
		return r == ',' || unicode.IsSpace(r)
	})
	if len(questions) == 0 || len(fields) != len(questions) {
		return quiz.Reply{}, false
	}
	answers := make([]quiz.Answer, 0, len(questions))
	for i, field := range fields {
		n, err := strconv.Atoi(field)
		if err != nil || n < 1 || n > 3 {
			return quiz.Reply{}, false
		}
		text := ""
		if n <= len(questions[i].Choices) {
			text = questions[i].Choices[n-1]
		}
		answers = append(answers, quiz.Answer{ID: questions[i].ID, Text: text, Choice: n})
	}
	return quiz.Reply{Attempted: true, Answers: answers}, true
}
