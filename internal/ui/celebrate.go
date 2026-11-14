package ui

import (
	"fmt"
	"io"
	"time"
)

var arrival = []string{
	"⛵  ≈≈≈≈≈≈  ⚓",
	"≈  ⛵  ≈≈≈≈  ⚓",
	"≈≈≈  ⛵  ≈≈  ⚓",
	"≈≈≈≈≈  ⛵  ≈  ⚓",
}

// celebratePass sails a boat into port after a passing quiz.
func celebratePass(w io.Writer) {
	if !interactive(w) {
		return
	}
	fmt.Fprint(w, "\033[?25l")
	for _, frame := range arrival {
		fmt.Fprintf(w, "\r%s  %s\033[K",
			paint(w, stampR, stampG, stampB, frame),
			paint(w, mutedR, mutedG, mutedB, "making port"),
		)
		time.Sleep(110 * time.Millisecond)
	}
	fmt.Fprint(w, "\r\033[K\033[?25h")
}
