// Package ui prints the terminal quiz and reads answers from the developer.
package ui

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"golang.org/x/term"

	"github.com/shipquiz/shipquiz/internal/quiz"
)

// PrintIntro shows the size, summary, and how many questions are coming.
func PrintIntro(out io.Writer, generated quiz.Generated) {
	label := "questions"
	if len(generated.Questions) == 1 {
		label = "question"
	}
	size := strings.TrimSpace(generated.Size)
	if size == "" {
		size = "change"
	}
	fmt.Fprintf(out, "\n%s\n", Stamp(out, "Change Size: "+titleWord(size)))
	fmt.Fprintln(out, Muted(out, fmt.Sprintf("%d %s", len(generated.Questions), label)))
	fmt.Fprintln(out, Stamp(out, "Diff Summary:"))
	writeWrapped(out, "", 0, generated.Summary)
}

// questionWidth is the longest quiz output line should run, in columns.
const questionWidth = 72

// CollectChoices reads one numbered choice per question and teaches immediately.
func CollectChoices(in *bufio.Reader, tty *os.File, out io.Writer, questions []quiz.Question) ([]quiz.Answer, error) {
	answers := make([]quiz.Answer, 0, len(questions))
	for i, question := range questions {
		writeQuestion(out, questionLabel(question.ID, i+1), question.Question)
		for choice, text := range question.Choices {
			label := fmt.Sprintf("   %d  ", choice+1)
			writeWrapped(out, Mark(out, label), len(label), text)
		}
		fmt.Fprintln(out, Muted(out, "   press 1, 2, or 3 · ? reveals the answer"))
		choice, err := readChoice(in, tty, out)
		if err != nil {
			return nil, err
		}
		showChoiceFeedback(out, question, choice)
		text := ""
		if choice > 0 && choice <= len(question.Choices) {
			text = question.Choices[choice-1]
		}
		answers = append(answers, quiz.Answer{ID: question.ID, Text: text, Choice: choice})
	}
	return answers, nil
}

func readChoice(in *bufio.Reader, tty *os.File, out io.Writer) (int, error) {
	fmt.Fprint(out, "   "+Stamp(out, ">")+" ")
	if tty != nil && term.IsTerminal(int(tty.Fd())) {
		state, err := term.MakeRaw(int(tty.Fd()))
		if err == nil {
			defer term.Restore(int(tty.Fd()), state)
			for {
				b, readErr := in.ReadByte()
				switch b {
				case '1', '2', '3':
					fmt.Fprintf(out, "%c\r\n", b)
					return int(b - '0'), nil
				case '?':
					fmt.Fprint(out, "?\r\n")
					return 0, nil
				case 3:
					fmt.Fprint(out, "\r\n")
					return 0, errors.New("quiz cancelled")
				case 4:
					fmt.Fprint(out, "\r\n")
					return 0, io.EOF
				}
				if readErr != nil {
					fmt.Fprint(out, "\r\n")
					return 0, readErr
				}
			}
		}
	}
	line, err := in.ReadString('\n')
	choice := strings.TrimSpace(line)
	if choice == "?" {
		return 0, err
	}
	if len(choice) == 1 && choice[0] >= '1' && choice[0] <= '3' {
		return int(choice[0] - '0'), err
	}
	if err != nil {
		return 0, err
	}
	return 0, errors.New("choose 1, 2, or 3")
}

func showChoiceFeedback(out io.Writer, question quiz.Question, choice int) {
	if choice == question.CorrectChoice {
		fmt.Fprintln(out, "   "+Pass(out, "✓ correct"))
	} else {
		answer := ""
		if question.CorrectChoice > 0 && question.CorrectChoice <= len(question.Choices) {
			answer = fmt.Sprintf("%d · %s", question.CorrectChoice, question.Choices[question.CorrectChoice-1])
		}
		label := "✗ not quite  "
		if choice == 0 {
			label = "answer  "
		}
		writeWrapped(out, "   "+Fail(out, label), 3+len([]rune(label)), answer)
	}
	if question.Explanation != "" {
		for _, line := range wrap(question.Explanation, questionWidth-3) {
			fmt.Fprintln(out, "   "+Muted(out, line))
		}
	}
}

func questionLabel(id string, fallback int) string {
	id = strings.TrimSpace(id)
	if len(id) > 1 && (id[0] == 'q' || id[0] == 'Q') {
		return "Q" + id[1:]
	}
	return fmt.Sprintf("Q%d", fallback)
}

// titleWord capitalizes the first letter of a size label.
func titleWord(word string) string {
	if word == "" {
		return word
	}
	return strings.ToUpper(word[:1]) + word[1:]
}

func writeQuestion(out io.Writer, label, question string) {
	fmt.Fprintln(out)
	writeWrapped(out, Mark(out, label)+"  ", len(label)+2, question)
}

func writeWrapped(out io.Writer, prefix string, prefixWidth int, text string) {
	if strings.TrimSpace(text) == "" {
		return
	}
	lines := wrap(text, questionWidth-prefixWidth)
	for i, line := range lines {
		if i == 0 {
			fmt.Fprintln(out, prefix+line)
			continue
		}
		fmt.Fprintln(out, strings.Repeat(" ", prefixWidth)+line)
	}
}

func wrap(text string, width int) []string {
	if width < 24 {
		width = 24
	}
	words := strings.Fields(text)
	if len(words) == 0 {
		return nil
	}
	var lines []string
	var cur string
	for _, word := range words {
		if cur == "" {
			cur = word
			continue
		}
		if len(cur)+1+len(word) > width {
			lines = append(lines, cur)
			cur = word
			continue
		}
		cur += " " + word
	}
	if cur != "" {
		lines = append(lines, cur)
	}
	return lines
}

// MissAction is the developer's choice after a failed quiz.
type MissAction int

const (
	// MissStop blocks the push. The developer cancelled the menu.
	MissStop MissAction = iota
	// MissPush allows the push after the developer has read the answers.
	MissPush
	// MissRetry asks only the questions they missed.
	MissRetry
	// MissRestart asks every question in this quiz again.
	MissRestart
	// MissNew generates a new quiz for the same diff.
	MissNew
)

// ChooseMissAction reads the next step after a failed quiz.
// allowRetry is set when at least one question was missed. allowNew is set when a model call is available.
func ChooseMissAction(in *bufio.Reader, tty *os.File, out io.Writer, allowRetry, allowNew bool) (MissAction, error) {
	fmt.Fprintln(out)
	fmt.Fprintln(out, Stamp(out, "What next?"))
	fmt.Fprintln(out, "   "+Mark(out, "1")+"  push anyway")
	if allowRetry {
		fmt.Fprintln(out, "   "+Mark(out, "2")+"  retry missed questions")
	}
	fmt.Fprintln(out, "   "+Mark(out, "3")+"  restart this quiz")
	if allowNew {
		fmt.Fprintln(out, "   "+Mark(out, "4")+"  generate new questions")
	}
	fmt.Fprint(out, "   "+Stamp(out, ">")+" ")
	action, err := readMissAction(in, tty, out, allowRetry, allowNew)
	return action, err
}

func readMissAction(in *bufio.Reader, tty *os.File, out io.Writer, allowRetry, allowNew bool) (MissAction, error) {
	if tty != nil && term.IsTerminal(int(tty.Fd())) {
		state, err := term.MakeRaw(int(tty.Fd()))
		if err == nil {
			defer term.Restore(int(tty.Fd()), state)
			for {
				b, readErr := in.ReadByte()
				if action, ok := missActionFromKey(b, allowRetry, allowNew); ok {
					fmt.Fprintf(out, "%c\r\n", b)
					return action, nil
				}
				switch b {
				case 3:
					fmt.Fprint(out, "\r\n")
					return MissStop, errors.New("quiz cancelled")
				case 4:
					fmt.Fprint(out, "\r\n")
					return MissStop, io.EOF
				}
				if readErr != nil {
					fmt.Fprint(out, "\r\n")
					return MissStop, readErr
				}
			}
		}
	}
	line, err := in.ReadString('\n')
	choice := strings.TrimSpace(line)
	if choice != "" {
		if action, ok := missActionFromKey(choice[0], allowRetry, allowNew); ok {
			return action, err
		}
	}
	if err != nil {
		return MissStop, err
	}
	return MissStop, errors.New("choose 1, 2, 3, or 4")
}

func missActionFromKey(b byte, allowRetry, allowNew bool) (MissAction, bool) {
	switch b {
	case '1':
		return MissPush, true
	case '2':
		if allowRetry {
			return MissRetry, true
		}
	case '3':
		return MissRestart, true
	case '4':
		if allowNew {
			return MissNew, true
		}
	}
	return MissStop, false
}

// PrintResult shows the grade after the local pass rule has been applied.
func PrintResult(out io.Writer, ev quiz.Evaluation) {
	tone := Fail
	if ev.Passed() {
		celebratePass(out)
		fmt.Fprintf(out, "\n%s  %s  %s\n",
			Pass(out, "⚓  PASS"),
			Pass(out, "cleared to ship! 🎉"),
			fmt.Sprintf("%.1f / 2", ev.Score),
		)
	} else {
		fmt.Fprintf(out, "\n%s  %s\n", tone(out, ev.Overall), fmt.Sprintf("%.1f / 2", ev.Score))
	}
	writeWrapped(out, "", 0, ev.Feedback)
	for _, item := range ev.PerQuestion {
		scoreTone := Fail
		if item.Score == 2 {
			scoreTone = Pass
		} else if item.Score == 1 {
			scoreTone = Muted
		}
		score := fmt.Sprintf("%d", item.Score)
		prefix := Mark(out, item.ID) + "  " + scoreTone(out, score) + "  "
		fmt.Fprintln(out)
		writeWrapped(out, prefix, len(item.ID)+len(score)+4, item.Comment)
	}
	fmt.Fprintln(out)
}
