package exception

import (
	"errors"
	"fmt"
)

var (
	ErrCommand  = errors.New("command err")
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
	allArgs := make([]any, 0, len(args)+1)
	allArgs = append(allArgs, kind)
	allArgs = append(allArgs, args...)

	return fmt.Errorf("%w: "+format, allArgs...)
}
