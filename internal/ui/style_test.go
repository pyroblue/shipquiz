package ui

import (
	"bytes"
	"strings"
	"testing"
)

func TestPlainWriterSkipsColor(t *testing.T) {
	var b bytes.Buffer
	Banner(&b)
	Section(&b, "Provider")
	Choice(&b, 1, "anthropic")
	if strings.Contains(b.String(), "\033") {
		t.Fatalf("color leaked into a pipe:\n%s", b.String())
	}
	if !strings.Contains(b.String(), "shipquiz") || !strings.Contains(b.String(), "anthropic") {
		t.Fatalf("%s", b.String())
	}
}

func TestHeroWordmarkFits(t *testing.T) {
	width := 0
	art := strings.Join(heroLines, "\n")
	for _, line := range heroLines {
		n := len([]rune(line))
		if n > heroWidth {
			t.Fatalf("len %d: %s", n, line)
		}
		if n > width {
			width = n
		}
	}
	if width != heroWidth {
		t.Fatalf("hero width %d", width)
	}
	if !strings.Contains(art, "█") || !strings.Contains(art, "▓") {
		t.Fatal("hero did not keep the wordmark")
	}
	var b bytes.Buffer
	Hero(&b)
	if strings.Contains(b.String(), "\033") {
		t.Fatalf("color leaked into a pipe:\n%s", b.String())
	}
	if !strings.Contains(b.String(), "You're the captain") {
		t.Fatalf("%s", b.String())
	}
}
