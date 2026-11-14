// Package ui prints the terminal quiz and reads answers from the developer.
package ui

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"github.com/shipquiz/shipquiz/internal/quiz"
)

// PrintIntro shows the size, summary, and how many questions are coming.
func PrintIntro(out io.Writer, generated quiz.Generated) {
	label := "questions"
	if len(generated.Questions) == 1 {
		label = "question"
	}
	size := generated.Size
	if size == "" {
		size = "change"
	}
	fmt.Fprintf(out, "\n%s · %d %s\n", size, len(generated.Questions), label)
	if generated.Summary != "" {
		fmt.Fprintln(out, generated.Summary)
	}
}

// CollectAnswers prompts for each question. A blank line submits the answer.
func CollectAnswers(in *bufio.Reader, out io.Writer, questions []quiz.Question) ([]quiz.Answer, error) {
	answers := make([]quiz.Answer, 0, len(questions))
	for i, question := range questions {
		fmt.Fprintf(out, "\n%d  %s\n", i+1, question.Question)
		if question.WhyItMatters != "" {
			fmt.Fprintf(out, "   %s\n", question.WhyItMatters)
		}
		fmt.Fprintln(out, "\n   answer (empty line to submit):")
		text, err := readBlock(in, out)
		if err != nil {
			return nil, err
		}
		answers = append(answers, quiz.Answer{ID: question.ID, Text: text})
	}
	return answers, nil
}

// PrintResult shows the grade after the local pass rule has been applied.
func PrintResult(out io.Writer, ev quiz.Evaluation) {
	fmt.Fprintf(out, "\n%s  %.1f / 2\n", ev.Overall, ev.Score)
	if ev.Feedback != "" {
		fmt.Fprintln(out, ev.Feedback)
	}
	for _, item := range ev.PerQuestion {
		fmt.Fprintf(out, "\n%s  %d  %s\n", item.ID, item.Score, item.Comment)
	}
	fmt.Fprintln(out)
}

func readBlock(in *bufio.Reader, out io.Writer) (string, error) {
	var lines []string
	for {
		fmt.Fprint(out, "   > ")
		line, err := in.ReadString('\n')
		trimmed := strings.TrimRight(line, "\r\n")
		if strings.TrimSpace(trimmed) == "" {
			if err != nil && len(lines) == 0 {
				return "", err
			}
			break
		}
		lines = append(lines, trimmed)
		if err != nil {
			break
		}
	}
	return strings.TrimSpace(strings.Join(lines, "\n")), nil
}
