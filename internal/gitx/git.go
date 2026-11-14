// Package gitx runs the git commands shipquiz needs to read a change.
package gitx

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// Output runs git and returns stdout.
// Stderr is included in the error so a missing repo or unknown revision is readable.
func Output(args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = err.Error()
		}
		return "", fmt.Errorf("git %s: %s", strings.Join(args, " "), msg)
	}
	return stdout.String(), nil
}

// Root returns the working tree root, or an error when cwd is not a repository.
func Root() (string, error) {
	out, err := Output("rev-parse", "--show-toplevel")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(out), nil
}

// Inside reports whether the current directory is inside a work tree.
func Inside() bool {
	_, err := Root()
	return err == nil
}
