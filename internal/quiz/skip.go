package quiz

import (
	"path/filepath"
	"strings"
)

var lockfiles = map[string]struct{}{
	"go.sum":            {},
	"package-lock.json": {},
	"pnpm-lock.yaml":    {},
	"yarn.lock":         {},
	"Cargo.lock":        {},
	"composer.lock":     {},
	"Gemfile.lock":      {},
	"poetry.lock":       {},
	"bun.lock":          {},
	"bun.lockb":         {},
	"uv.lock":           {},
	"flake.lock":        {},
	"shrinkwrap.yaml":   {},
	"Package.resolved":  {},
}

// ShouldSkip reports whether every changed file is a lockfile or matches skipPaths.
// An empty file list is not a skip: the caller treats an empty diff separately.
func ShouldSkip(files, skipPaths []string) bool {
	if len(files) == 0 {
		return false
	}
	for _, file := range files {
		if !skippable(file, skipPaths) {
			return false
		}
	}
	return true
}

func skippable(file string, patterns []string) bool {
	base := filepath.Base(filepath.ToSlash(file))
	if _, ok := lockfiles[base]; ok {
		return true
	}
	for _, pattern := range patterns {
		if MatchPath(pattern, file) {
			return true
		}
	}
	return false
}

// MatchPath matches a shipquiz skip pattern against a repo-relative path.
// Patterns support *.md, docs/**, and **/*_test.go.
func MatchPath(pattern, path string) bool {
	pattern = filepath.ToSlash(strings.TrimSpace(pattern))
	path = filepath.ToSlash(path)
	if pattern == "" {
		return false
	}
	if strings.HasPrefix(pattern, "**/") {
		rest := strings.TrimPrefix(pattern, "**/")
		if glob(rest, path) || glob(rest, filepath.Base(path)) {
			return true
		}
	}
	if strings.HasSuffix(pattern, "/**") {
		prefix := strings.TrimSuffix(pattern, "/**")
		return path == prefix || strings.HasPrefix(path, prefix+"/")
	}
	if strings.Contains(pattern, "/") {
		return glob(pattern, path)
	}
	return glob(pattern, filepath.Base(path))
}

func glob(pattern, path string) bool {
	ok, err := filepath.Match(pattern, path)
	return err == nil && ok
}

// SubstantiveLines counts added and removed diff lines, ignoring headers.
func SubstantiveLines(diff string) int {
	n := 0
	for _, line := range strings.Split(diff, "\n") {
		if strings.HasPrefix(line, "+++") || strings.HasPrefix(line, "---") {
			continue
		}
		if strings.HasPrefix(line, "+") || strings.HasPrefix(line, "-") {
			n++
		}
	}
	return n
}
