package exception

import (
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/serenitysz/serenity/internal/render"
)

var (
	ErrCommand  = errors.New("command failed")
	ErrInternal = errors.New("internal error")
)

func ExitCode(err error) int {
	switch {
	case err == nil:
		return 0
	case errors.Is(err, ErrCommand):
		return 1
	default:
		return 2
	}
}

func CommandError(format string, args ...any) error {
	return wrap(ErrCommand, format, args...)
}

func InternalError(format string, args ...any) error {
	return wrap(ErrInternal, format, args...)
}

func wrap(kind error, format string, args ...any) error {
	if existing := passthroughWrappedError(format, args...); existing != nil {
		return existing
	}

	allArgs := make([]any, 0, len(args)+1)
	allArgs = append(allArgs, kind)
	allArgs = append(allArgs, args...)

	return fmt.Errorf("%w: "+format, allArgs...)
}

func Message(err error) string {
	if err == nil {
		return ""
	}

	msg := err.Error()

	for {
		trimmed := strings.TrimPrefix(msg, ErrCommand.Error()+": ")
		if trimmed != msg {
			msg = trimmed
			continue
		}

		trimmed = strings.TrimPrefix(msg, ErrInternal.Error()+": ")
		if trimmed != msg {
			msg = trimmed
			continue
		}

		break
	}

	return msg
}

func Write(w io.Writer, err error, noColor bool) {
	if err == nil {
		return
	}

	label := "error:"
	color := render.Red

	switch {
	case errors.Is(err, ErrInternal):
		label = "internal error:"
		color = render.Red
	case errors.Is(err, ErrCommand):
		label = "failed:"
		color = render.Yellow
	}

	_, _ = fmt.Fprintf(w, "%s %s\n", render.Paint(label, color, noColor), Message(err))
}

func passthroughWrappedError(format string, args ...any) error {
	if len(args) != 1 {
		return nil
	}

	err, ok := args[0].(error)
	if !ok || err == nil {
		return nil
	}

	switch format {
	case "%v", "%w", "%s":
		if errors.Is(err, ErrCommand) || errors.Is(err, ErrInternal) {
			return err
		}
	}

	return nil
}
