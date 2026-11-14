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
	}
	encoded, err := encodeMarker(marker)
	if err != nil {
		return "", err
	}
	var b strings.Builder
	fmt.Fprintf(&b, "<!-- shipquiz %s -->\n\n", encoded)
	fmt.Fprintf(&b, "## shipquiz\n\n**%s** — %s\n\n", blank(marker.Quiz.Size, "change"), marker.Quiz.Summary)
	b.WriteString("Reply in this thread with your answers, in your own words. The model matches the reply to each question.\n\n")
	for _, question := range marker.Quiz.Questions {
		fmt.Fprintf(&b, "### %s\n\n%s\n\n", question.ID, question.Question)
	}
	return b.String(), nil
}

func renderResult(ev quiz.Evaluation) string {
	var b strings.Builder
	b.WriteString("<!-- shipquiz result -->\n\n")
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
