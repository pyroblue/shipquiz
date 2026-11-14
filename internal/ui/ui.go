// Package ui prints the terminal quiz and reads answers from the developer.
package ui

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"
	"unicode/utf8"

	"golang.org/x/term"

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
	fmt.Fprintf(out, "\n%s  %s\n", Stamp(out, size), Muted(out, fmt.Sprintf("· %d %s", len(generated.Questions), label)))
	writeWrapped(out, "", 0, generated.Summary)
}

// questionWidth is the longest quiz output line should run, in columns.
const questionWidth = 72

// CollectAnswers prompts for each question. A blank line submits the answer.
// Ctrl-G prints the reference answer. tty is the terminal, when there is one.
func CollectAnswers(in *bufio.Reader, tty *os.File, out io.Writer, questions []quiz.Question) ([]quiz.Answer, error) {
	answers := make([]quiz.Answer, 0, len(questions))
	for i, question := range questions {
		writeQuestion(out, questionLabel(question.ID, i+1), question.Question)
		fmt.Fprintln(out, Muted(out, "   ctrl-g shows the answer · blank line submits"))
		text, err := readAnswer(in, tty, out, question.Reference)
		if err != nil {
			return nil, err
		}
		answers = append(answers, quiz.Answer{ID: question.ID, Text: text})
	}
	return answers, nil
}

func questionLabel(id string, fallback int) string {
	id = strings.TrimSpace(id)
	if len(id) > 1 && (id[0] == 'q' || id[0] == 'Q') {
		return id[1:]
	}
	return fmt.Sprintf("%d", fallback)
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

func readAnswer(in *bufio.Reader, tty *os.File, out io.Writer, reference string) (string, error) {
	if tty != nil && term.IsTerminal(int(tty.Fd())) {
		state, err := term.MakeRaw(int(tty.Fd()))
		if err == nil {
			defer term.Restore(int(tty.Fd()), state)
			return editAnswer(in, tty, out, reference, true)
		}
	}
	return readBlock(in, out)
}

func editAnswer(in *bufio.Reader, tty *os.File, out io.Writer, reference string, raw bool) (string, error) {
	var lines []string
	var cur []rune
	nl := func() {
		if raw {
			fmt.Fprint(out, "\r\n")
			return
		}
		fmt.Fprintln(out)
	}
	prompt := func() {
		fmt.Fprint(out, "   "+Stamp(out, ">")+" ")
		fmt.Fprint(out, string(cur))
	}
	redraw := func() {
		fmt.Fprint(out, "\r\033[K")
		prompt()
	}
	prompt()
	for {
		b, err := in.ReadByte()
		if err != nil {
			nl()
			if len(lines) == 0 && len(cur) == 0 {
				return "", err
			}
			lines = append(lines, string(cur))
			return strings.TrimSpace(strings.Join(lines, "\n")), nil
		}
		switch b {
		case 3:
			nl()
			return "", errors.New("quiz cancelled")
		case 4:
			if len(cur) == 0 && len(lines) == 0 {
				nl()
				return "", io.EOF
			}
		case 7:
			nl()
			showReference(out, reference, raw)
			prompt()
		case 8, 127:
			if len(cur) > 0 {
				cur = cur[:len(cur)-1]
				redraw()
			}
		case '\r', '\n':
			if len(cur) == 0 {
				nl()
				return strings.TrimSpace(strings.Join(lines, "\n")), nil
			}
			lines = append(lines, string(cur))
			cur = nil
			nl()
			prompt()
		case 21:
			cur = nil
			redraw()
		case 27:
			swallowEscape(in, tty)
		default:
			if b < 32 {
				continue
			}
			r, err := readRune(in, b)
			if err != nil {
				nl()
				return "", err
			}
			cur = append(cur, r)
			fmt.Fprint(out, string(r))
		}
	}
}

func showReference(out io.Writer, reference string, raw bool) {
	end := "\n"
	if raw {
		end = "\r\n"
	}
	text := strings.TrimSpace(reference)
	if text == "" {
		fmt.Fprint(out, "   "+Muted(out, "no answer stored for this question")+end)
		return
	}
	const prefixWidth = 11 // "   answer  "
	for i, line := range wrap(text, questionWidth-prefixWidth) {
		if i == 0 {
			fmt.Fprint(out, "   "+Stamp(out, "answer")+"  "+line+end)
			continue
		}
		fmt.Fprint(out, "           "+line+end)
	}
}

func readRune(in *bufio.Reader, first byte) (rune, error) {
	if first < 128 {
		return rune(first), nil
	}
	need := 1
	switch {
	case first&0xE0 == 0xC0:
		need = 2
	case first&0xF0 == 0xE0:
		need = 3
	case first&0xF8 == 0xF0:
		need = 4
	default:
		return utf8.RuneError, nil
	}
	buf := make([]byte, need)
	buf[0] = first
	for i := 1; i < need; i++ {
		b, err := in.ReadByte()
		if err != nil {
			return utf8.RuneError, err
		}
		buf[i] = b
	}
	r, _ := utf8.DecodeRune(buf)
	return r, nil
}

func swallowEscape(in *bufio.Reader, tty *os.File) {
	if tty != nil {
		_ = tty.SetReadDeadline(time.Now().Add(20 * time.Millisecond))
		defer tty.SetReadDeadline(time.Time{})
	}
	for {
		if tty == nil && in.Buffered() == 0 {
			return
		}
		b, err := in.ReadByte()
		if err != nil {
			return
		}
		if (b >= 'A' && b <= 'Z') || (b >= 'a' && b <= 'z') || b == '~' {
			return
		}
	}
}

func readBlock(in *bufio.Reader, out io.Writer) (string, error) {
	var lines []string
	for {
		fmt.Fprint(out, "   "+Stamp(out, ">")+" ")
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
