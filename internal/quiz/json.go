package quiz

import (
	"encoding/json"
	"fmt"
	"strings"
)

// ParseGenerated extracts the quiz JSON from a model response.
func ParseGenerated(raw string) (Generated, error) {
	var g Generated
	if err := json.Unmarshal([]byte(extractJSON(raw)), &g); err != nil {
		return Generated{}, fmt.Errorf("could not parse quiz JSON: %w", err)
	}
	return g, nil
}

// ParseEvaluation extracts the grade JSON from a model response.
func ParseEvaluation(raw string) (Evaluation, error) {
	var ev Evaluation
	if err := json.Unmarshal([]byte(extractJSON(raw)), &ev); err != nil {
		return Evaluation{}, fmt.Errorf("could not parse grade JSON: %w", err)
	}
	return ev, nil
}

func extractJSON(raw string) string {
	s := strings.TrimSpace(raw)
	if start := strings.Index(s, "```"); start >= 0 {
		rest := s[start+3:]
		rest = strings.TrimPrefix(rest, "json")
		rest = strings.TrimPrefix(rest, "JSON")
		if end := strings.Index(rest, "```"); end >= 0 {
			s = rest[:end]
		} else {
			s = rest
		}
	}
	s = strings.TrimSpace(s)
	from := strings.Index(s, "{")
	to := strings.LastIndex(s, "}")
	if from >= 0 && to > from {
		return s[from : to+1]
	}
	return s
}
