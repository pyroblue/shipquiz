package ui

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"golang.org/x/term"
)

// wake is a boat crossing one wave. Each frame is the same width when ⛵ is wide.
var wake = []string{
	"≈≈⛵",
	"≈⛵≈",
	"⛵≈≈",
	"≈⛵≈",
}

// bearings are the lines that rotate beside the task.
var bearings = []string{
	"reading the chart",
	"sounding the depth",
	"plotting a course",
	"watching the wake",
}

type spinner struct {
	w     io.Writer
	label string
	stop  chan struct{}
	done  chan struct{}
	once  sync.Once
}

// Spin shows a nautical loader until the returned function is called.
// A pipe gets one plain line, with no animation.
func Spin(w io.Writer, label string) func() {
	if !interactive(w) {
		fmt.Fprintf(w, "shipquiz: %s…\n", label)
		return func() {}
	}
	s := &spinner{
		w:     w,
		label: label,
		stop:  make(chan struct{}),
		done:  make(chan struct{}),
	}
	fmt.Fprint(w, "\033[?25l")
	go s.run()
	return s.Stop
}

func (s *spinner) run() {
	defer close(s.done)
	tick := time.NewTicker(140 * time.Millisecond)
	defer tick.Stop()
	frame := 0
	s.draw(frame)
	for {
		select {
		case <-s.stop:
			fmt.Fprint(s.w, "\r\033[K\033[?25h")
			return
		case <-tick.C:
			frame++
			s.draw(frame)
		}
	}
}

func (s *spinner) draw(frame int) {
	boat := wake[frame%len(wake)]
	bearing := bearings[(frame/4)%len(bearings)]
	fmt.Fprintf(s.w, "\rshipquiz  %s  %s  %s\033[K",
		paint(s.w, stampR, stampG, stampB, boat),
		s.label,
		paint(s.w, mutedR, mutedG, mutedB, "· "+bearing),
	)
}

// Stop ends the loader. A second call does nothing.
func (s *spinner) Stop() {
	s.once.Do(func() { close(s.stop) })
	<-s.done
}

func interactive(w io.Writer) bool {
	if os.Getenv("TERM") == "dumb" {
		return false
	}
	file, ok := w.(*os.File)
	if !ok {
		return false
	}
	return term.IsTerminal(int(file.Fd()))
}
