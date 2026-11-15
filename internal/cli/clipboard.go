package cli

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

// copyText puts text on the system clipboard.
func copyText(text string) error {
	name, args, err := clipboardCommand()
	if err != nil {
		return err
	}
	cmd := exec.Command(name, args...)
	cmd.Stdin = strings.NewReader(text)
	return cmd.Run()
}

// clipboardCommand is the platform tool that reads the clipboard from stdin.
func clipboardCommand() (string, []string, error) {
	switch runtime.GOOS {
	case "darwin":
		return "pbcopy", nil, nil
	case "linux":
		if _, err := exec.LookPath("wl-copy"); err == nil {
			return "wl-copy", nil, nil
		}
		if _, err := exec.LookPath("xclip"); err == nil {
			return "xclip", []string{"-selection", "clipboard"}, nil
		}
		return "", nil, fmt.Errorf("install wl-copy or xclip to copy from the terminal")
	default:
		return "", nil, fmt.Errorf("clipboard is not available on %s", runtime.GOOS)
	}
}
