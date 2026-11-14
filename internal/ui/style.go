package ui

import (
	"fmt"
	"io"
	"os"
	"strconv"

	"golang.org/x/term"
)

// Colors are the site palette, lifted so they read on a dark terminal.
// stamp is the teal wordmark, mark is the steel number, muted is the aside.
const (
	stampR, stampG, stampB = 58, 148, 168
	markR, markG, markB    = 126, 184, 206
	mutedR, mutedG, mutedB = 138, 154, 166
	passR, passG, passB    = 62, 168, 112
	failR, failG, failB    = 196, 112, 96
)

// Banner prints the wordmark and the line under it.
func Banner(w io.Writer) {
	fmt.Fprintln(w)
	fmt.Fprintln(w, bold(w, stampR, stampG, stampB, "shipquiz"))
	fmt.Fprintln(w, paint(w, mutedR, mutedG, mutedB, "You're the captain. Ship only what you can explain."))
}

// heroWidth is the widest line of the install wordmark, in columns.
const heroWidth = 150

// heroLines is the SHIPQUIZ wordmark shown the first time someone installs.
var heroLines = []string{
	`                            ▒░              ░░░░       ░░                                 ░░                     ▒▒▒▒▒░    ░░          ░     ░▒▒▒▒▒▒▒`,
	`                  ▒███░ ▓████▓  ▓███████▓  ▒████████ ▒████████   ▒█████████████▓      ▓███▓░▓███    ▒████████▒  ▓█████▓  ░████████░   ▓██████████████▓`,
	`                 ▓██▓     ▒██░    ████▒      █████     ▓████       ▓████▒   ▓███▓  ▒████▓     ▓███    ▓████░      ▓█▓      ▒███▓     ░██▓     ▓█████`,
	`                 ████▓     ▓▓    ▓███▓      ▒████░     ████        ▓███▓    ▓███▓ ▒████       ░███▓  ▒████▒       ██░      ████░     ░▓▒    ▒█████░`,
	`                 ░█████▒        ▒████▒      ████▒     ▓███░       ▒████   ▒████▒ ▓████        ▒████▒ ▓███▓       ▓█▒      ▓███▒           ░█████▒`,
	`                   ▒█████░      ▓██████████████▓      ███▓       ░█████████▓░░   ████         ▓████▒▒████        ██       ███▓          ░█████▓`,
	`               ▒░    ▓████▒    ▒████      ░████░     ████░      ░████▒           ████        ▓████▓░████░       ██       ▓███▒         ▓█████`,
	`              ▒██     ▒████   ░████░      ▓███▓     ▓███▒       █████            ████       ▒████▓ ░████       ▓█▒      ▓███▓        ▒█████▒      ▓▓`,
	`              ▒███▒   ▒███▒   ▓████      ▓████░    ▓████▒      ░█████            ░████     ▓███▓   ░█████     ▓█▓      ▓████▒       ██████░    ░▓██░`,
	`              ▓█▒▒▓█████▒  ░████████▓ ░████████▓░▓███████▒   █████████▒            ░▓███████▓       ░████████▓▒      ▒███████▓    ░███████████████░`,
	`                            ░░░░░                                                    ▓███▒░             ░░`,
	`                                                                                   ▒██████████▓▓▒▒▒▓██░`,
	`                                                                                  ▒█▒     ░▓████████▒`,
}

// Hero prints the install wordmark. A narrow terminal gets the small banner instead.
func Hero(w io.Writer) {
	if termWidth(w) < heroWidth {
		Banner(w)
		return
	}
	fmt.Fprintln(w)
	for _, line := range heroLines {
		fmt.Fprintln(w, paint(w, stampR, stampG, stampB, line))
	}
	fmt.Fprintln(w, paint(w, mutedR, mutedG, mutedB, "You're the captain. Ship only what you can explain."))
}

func termWidth(w io.Writer) int {
	file, ok := w.(*os.File)
	if !ok {
		return 80
	}
	width, _, err := term.GetSize(int(file.Fd()))
	if err != nil || width <= 0 {
		return 80
	}
	return width
}

// Section prints a teal heading.
func Section(w io.Writer, title string) {
	fmt.Fprintln(w)
	fmt.Fprintln(w, paint(w, stampR, stampG, stampB, title))
}

// Choice prints one numbered row. The number uses the steel accent.
func Choice(w io.Writer, n int, label string) {
	fmt.Fprintf(w, "  %s  %s\n", paint(w, markR, markG, markB, fmt.Sprintf("%2d", n)), label)
}

// Hint prints an aside in the muted tone.
func Hint(w io.Writer, text string) {
	fmt.Fprintln(w, paint(w, mutedR, mutedG, mutedB, text))
}

// Ask builds a prompt. The label is teal and the default sits in brackets.
func Ask(w io.Writer, label, def string) string {
	if def == "" {
		return paint(w, stampR, stampG, stampB, label) + ": "
	}
	return paint(w, stampR, stampG, stampB, label) + " " + paint(w, mutedR, mutedG, mutedB, "["+def+"]") + ": "
}

// Stamp colors a short label in the teal accent.
func Stamp(w io.Writer, text string) string {
	return paint(w, stampR, stampG, stampB, text)
}

// Muted colors secondary text.
func Muted(w io.Writer, text string) string {
	return paint(w, mutedR, mutedG, mutedB, text)
}

// Mark colors a number or id.
func Mark(w io.Writer, text string) string {
	return paint(w, markR, markG, markB, text)
}

// Pass colors a successful result.
func Pass(w io.Writer, text string) string {
	return paint(w, passR, passG, passB, text)
}

// Fail colors a failed result.
func Fail(w io.Writer, text string) string {
	return paint(w, failR, failG, failB, text)
}

// Row prints a status line with a muted label.
func Row(w io.Writer, label, value string) {
	fmt.Fprintf(w, "  %s  %s\n", paint(w, mutedR, mutedG, mutedB, fmt.Sprintf("%-8s", label)), value)
}

// Command prints a command name and what it does.
func Command(w io.Writer, name, detail string) {
	fmt.Fprintf(w, "  %s  %s\n", paint(w, stampR, stampG, stampB, fmt.Sprintf("%-16s", name)), paint(w, mutedR, mutedG, mutedB, detail))
}

func paint(w io.Writer, r, g, b int, text string) string {
	if text == "" || !colorOn(w) {
		return text
	}
	return "\033[38;2;" + strconv.Itoa(r) + ";" + strconv.Itoa(g) + ";" + strconv.Itoa(b) + "m" + text + "\033[0m"
}

func bold(w io.Writer, r, g, b int, text string) string {
	if text == "" || !colorOn(w) {
		return text
	}
	return "\033[1;38;2;" + strconv.Itoa(r) + ";" + strconv.Itoa(g) + ";" + strconv.Itoa(b) + "m" + text + "\033[0m"
}

func colorOn(w io.Writer) bool {
	if _, ok := os.LookupEnv("NO_COLOR"); ok {
		return false
	}
	if os.Getenv("TERM") == "dumb" {
		return false
	}
	file, ok := w.(*os.File)
	if !ok {
		return false
	}
	return term.IsTerminal(int(file.Fd()))
}
