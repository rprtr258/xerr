package xerr

import (
	"errors"
	"fmt"
	"runtime"
	"strconv"
	"strings"
)

const _maxStackDepth = 50

type stackFrame struct {
	function string
	file     string
	line     int
}

type xError struct {
	// err guranteed to be not nil
	err   error
	stack []stackFrame
}

func (err xError) Error() string {
	if err.stack == nil {
		return err.Error()
	}

	sb := strings.Builder{}
	sb.WriteString(err.err.Error())
	sb.WriteString("\nStacktrace:\n")
	for _, frame := range err.stack {
		sb.WriteString(frame.function)
		sb.WriteString("\n\t")
		sb.WriteString(frame.file)
		sb.WriteString(":")
		sb.WriteString(strconv.Itoa(frame.line))
		sb.WriteString("\n")
	}
	return sb.String()
}

func stacktrace() []stackFrame {
	callers := make([]uintptr, _maxStackDepth)
	length := runtime.Callers(2, callers[:])
	callers = callers[:length]

	frames := runtime.CallersFrames(callers)
	stack := make([]stackFrame, 0, len(callers))
	for {
		frame, more := frames.Next()
		stack = append(
			stack,
			stackFrame{
				function: frame.Function,
				file:     frame.File,
				line:     frame.Line,
			},
		)
		if !more {
			break
		}
	}
	return stack
}

type option func()

func WithWrap() option { return nil }

func WithStack() option { return nil }

func WithNoTimestamp() option { return nil }

func WithMessage(message string) option { return nil }

func WithField[T any](name string, value T) option { return nil }

func New(...option) error {
	return fmt.Errorf(format, args...)
}

func Wrap(err error, format string, args ...any) error {
	message := fmt.Sprintf(format, args...)
	return fmt.Errorf("%s: %w", message, err)
}

func NewST(format string, args ...any) error {
	return xError{
		err:   fmt.Errorf(format, args...),
		stack: stacktrace(),
	}
}

func WrapST(err error, format string, args ...any) error {
	message := fmt.Sprintf(format, args...)
	return xError{
		err:   fmt.Errorf("%s: %w", message, err),
		stack: stacktrace(),
	}
}

func Unwrap(err error) error {
	return errors.Unwrap(err)
}

func Unwraps(err error) []error {
	if e, ok := err.(interface {
		Unwrap() []error
	}); ok {
		return e.Unwrap()
	}

	if res := Unwrap(err); res != nil {
		return []error{res}
	}

	return nil
}

func Is(err, target error) bool {
	return errors.Is(err, target)
}

func As[E error](err error) (E, bool) {
	var res E
	ok := errors.As(err, &res)
	return res, ok
}

func Errors(err error) []error {
	if errs, ok := As[multiError](err); ok {
		return errs.errs
	}

	if err != nil {
		return []error{err}
	}

	return nil
}
