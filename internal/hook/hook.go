// Package hook installs the shipquiz pre-push hook into the current repository.
package hook

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/shipquiz/shipquiz/internal/gitx"
)

const marker = "# shipquiz-managed"

// Script is the pre-push hook. It forwards git's stdin, then any earlier hook.
func Script() string {
	return `#!/bin/sh
# shipquiz-managed
set -u
if [ -n "${SHIPQUIZ_SKIP:-}" ]; then
  echo "shipquiz: skipped (SHIPQUIZ_SKIP is set)"
  exit 0
fi
if ! command -v shipquiz >/dev/null 2>&1; then
  echo "shipquiz: not found on PATH. Install it, or push with --no-verify." >&2
  exit 1
fi
hook_dir=$(CDPATH= cd -- "$(dirname "$0")" && pwd)
input=$(cat)
printf '%s\n' "$input" | shipquiz hook-run
status=$?
if [ -x "$hook_dir/pre-push.previous" ]; then
  printf '%s\n' "$input" | "$hook_dir/pre-push.previous"
  prev=$?
  if [ "$status" -eq 0 ]; then
    status=$prev
  fi
fi
exit "$status"
`
}

// Install writes the pre-push hook, preserving an existing unrelated hook.
func Install() error {
	dir, err := hooksDir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	path := filepath.Join(dir, "pre-push")
	existing, err := os.ReadFile(path)
	if err == nil && !bytes.Contains(existing, []byte(marker)) {
		if err := os.WriteFile(filepath.Join(dir, "pre-push.previous"), existing, 0o755); err != nil {
			return err
		}
	} else if err != nil && !os.IsNotExist(err) {
		return err
	}
	return os.WriteFile(path, []byte(Script()), 0o755)
}

// Uninstall removes a shipquiz hook and restores a previous one if we saved it.
func Uninstall() error {
	dir, err := hooksDir()
	if err != nil {
		return err
	}
	path := filepath.Join(dir, "pre-push")
	existing, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	if !bytes.Contains(existing, []byte(marker)) {
		return fmt.Errorf("%s is not a shipquiz hook", path)
	}
	if err := os.Remove(path); err != nil {
		return err
	}
	prev := filepath.Join(dir, "pre-push.previous")
	if _, err := os.Stat(prev); err == nil {
		return os.Rename(prev, path)
	}
	return nil
}

// Installed reports whether the current repo's pre-push hook is shipquiz's.
func Installed() bool {
	dir, err := hooksDir()
	if err != nil {
		return false
	}
	existing, err := os.ReadFile(filepath.Join(dir, "pre-push"))
	return err == nil && bytes.Contains(existing, []byte(marker))
}

func hooksDir() (string, error) {
	out, err := gitx.Output("rev-parse", "--git-path", "hooks")
	if err != nil {
		return "", fmt.Errorf("run this inside a git repository")
	}
	path := strings.TrimSpace(out)
	if !filepath.IsAbs(path) {
		cwd, err := os.Getwd()
		if err != nil {
			return "", err
		}
		path = filepath.Join(cwd, path)
	}
	return path, nil
}
