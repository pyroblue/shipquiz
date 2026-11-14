package cli

import (
	"bufio"
	"fmt"
	"io"
	"os"

	"golang.org/x/term"
)

type console struct {
	in   *bufio.Reader
	out  io.Writer
	file *os.File
}

func openConsole() console {
	file, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
	if err != nil {
		return console{in: bufio.NewReader(os.Stdin), out: os.Stdout}
	}
	return console{
		in:   bufio.NewReader(file),
		out:  file,
		file: file,
	}
}

func (c console) close() {
	if c.file != nil {
		_ = c.file.Close()
	}
}

func (c console) prompt(label string) (string, error) {
	fmt.Fprint(c.out, label)
	line, err := c.in.ReadString('\n')
	return trimLine(line), err
}

func (c console) confirm(label string, def bool) (bool, error) {
	suffix := " [y/N] "
	if def {
		suffix = " [Y/n] "
	}
	answer, err := c.prompt(label + suffix)
	if answer == "" {
		return def, err
	}
	switch answer {
	case "y", "Y", "yes", "YES":
		return true, err
	default:
		return false, err
	}
}

func (c console) secret(label string) (string, error) {
	fmt.Fprint(c.out, label)
	if c.file == nil || !term.IsTerminal(int(c.file.Fd())) {
		fmt.Fprintln(c.out)
		return "", fmt.Errorf("setup needs a terminal to read the API key")
	}
	raw, err := term.ReadPassword(int(c.file.Fd()))
	fmt.Fprintln(c.out)
	if err != nil {
		return "", err
	}
	return trimLine(string(raw)), nil
}

func trimLine(line string) string {
	for len(line) > 0 && (line[len(line)-1] == '\n' || line[len(line)-1] == '\r' || line[len(line)-1] == ' ' || line[len(line)-1] == '\t') {
		line = line[:len(line)-1]
	}
	start := 0
	for start < len(line) && (line[start] == ' ' || line[start] == '\t') {
		start++
	}
	return line[start:]
}
