package check

import (
	"errors"
	"testing"

	"github.com/serenitysz/serenity/internal/exception"
)

func TestValidateOptionsRejectsUnsafeWithoutWrite(t *testing.T) {
	t.Parallel()

	err := validateOptions(&CheckOptions{Unsafe: true})
	if !errors.Is(err, exception.ErrCommand) {
		t.Fatalf("expected command error, got %v", err)
	}

	if got := exception.Message(err); got != "--unsafe requires --write" {
		t.Fatalf("unexpected validation error: %q", got)
	}
}
