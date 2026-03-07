package render

import (
	"bytes"
	"testing"
)

func TestLogfRendersTaggedLineWithoutColor(t *testing.T) {
	var buf bytes.Buffer

	SetNoColor(true)
	defer SetNoColor(false)

	Logf(&buf, "done", Green, "updated %s", "serenity")

	if got := buf.String(); got != "done  updated serenity\n" {
		t.Fatalf("unexpected rendered log line: %q", got)
	}
}

func TestTagPadsLabelWithoutColor(t *testing.T) {
	SetNoColor(true)
	defer SetNoColor(false)

	if got := Tag("info", Blue, false); got != "info " {
		t.Fatalf("unexpected tag: %q", got)
	}
}
