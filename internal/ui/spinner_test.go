package ui

import (
	"bytes"
	"strings"
	"testing"
)

func TestSpinOnAPipeStaysPlain(t *testing.T) {
	var b bytes.Buffer
	stop := Spin(&b, "generating questions")
	stop()
	got := b.String()
	if strings.Contains(got, "\r") || strings.Contains(got, "\033") {
		t.Fatalf("animation leaked into a pipe:\n%q", got)
	}
	if !strings.Contains(got, "generating questions") {
		t.Fatalf("%q", got)
	}
}
