package exception

import (
	"bytes"
	"errors"
	"testing"
)

func TestInternalErrorPassthroughKeepsExistingKind(t *testing.T) {
	t.Parallel()

	original := CommandError("found 2 warnings")
	wrapped := InternalError("%v", original)

	if !errors.Is(wrapped, ErrCommand) {
		t.Fatalf("expected command error to be preserved, got %v", wrapped)
	}

	if errors.Is(wrapped, ErrInternal) {
		t.Fatalf("did not expect internal error wrapping, got %v", wrapped)
	}
}

func TestMessageStripsNestedPrefixes(t *testing.T) {
	t.Parallel()

	err := InternalError("%w", InternalError("could not read config"))

	if got := Message(err); got != "could not read config" {
		t.Fatalf("expected stripped message, got %q", got)
	}
}

func TestWriteFormatsCommandError(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer

	Write(&buf, CommandError("found 3 warnings"), true)

	if got := buf.String(); got != "failed: found 3 warnings\n" {
		t.Fatalf("unexpected rendered command error: %q", got)
	}
}

func TestWriteFormatsInternalError(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer

	Write(&buf, InternalError("could not parse config file %q", "serenity.json"), true)

	if got := buf.String(); got != "internal error: could not parse config file \"serenity.json\"\n" {
		t.Fatalf("unexpected rendered internal error: %q", got)
	}
}
