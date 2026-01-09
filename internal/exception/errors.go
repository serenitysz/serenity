package exception

import (
	"errors"
	"fmt"
)

var (
	ErrCheck    = errors.New("check failed")
	ErrInternal = errors.New("internal error")
)

func ExitCode(err error) int {
	switch {
	case err == nil:
		return 0
	case errors.Is(err, ErrCheck):
		return 1
	default:
		return 2
	}
}

func CheckError(format string, args ...any) error {
	return fmt.Errorf("%w: %s", ErrCheck, fmt.Sprintf(format, args...))
}

func InternalError(format string, args ...any) error {
	return fmt.Errorf("%w: %s", ErrInternal, fmt.Sprintf(format, args...))
}
