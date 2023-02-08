package xerr

import (
	"errors"
	"fmt"
	"runtime"
	"strings"
	"time"
)

const _maxStackDepth = 50

type stackFrame struct {
	function string
	file     string
	line     int
}

type xError struct {
	// errs guranteed to be not nil
	errs    []error
	stack   []stackFrame
	message string
	fields  map[string]any
	at      time.Time
}

func (err *xError) timestamp() string {
	if err.at.IsZero() {
		return ""
	}
	return err.at.String()
}

func (err *xError) errors() string {
	errMessages := make([]string, len(err.errs))
	for i, err := range err.errs {
		errMessages[i] = err.Error()
	}

	return strings.Join(errMessages, ";")
}

func (err *xError) Error() string {
	if err.stack != nil {
		frames := make([]map[string]any, len(err.stack))
		for i, frame := range err.stack {
			frames[i] = map[string]any{
				"function": frame.function,
				"file":     frame.file,
				"line":     frame.line,
			}
		}
		err.fields["stacktrace"] = frames
	}

	if err.message != "" {
		err.fields["message"] = err.message
	}

	err.fields["errors"] = err.errors()
	err.fields["at"] = err.timestamp()

	return fmt.Sprint(err.fields)
}

func (err *xError) Unwrap() error {
	if len(err.errs) == 0 {
		return nil
	}

	return err.errs[0]
}

func (err *xError) Unwraps() []error {
	return err.errs
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

type option func(*xError)

func WithErr(err error) option {
	return func(xe *xError) {
		xe.errs = append(xe.errs, err)
	}
}

func WithStack() option {
	return func(xe *xError) {
		xe.stack = stacktrace()
	}
}

func WithMessage(message string) option {
	return func(xe *xError) {
		xe.message = message
	}
}

func WithField[T any](name string, value T) option {
	return func(xe *xError) {
		if xe.fields == nil {
			xe.fields = map[string]any{
				name: value,
			}
		} else {
			xe.fields[name] = value
		}
	}
}

func New(options ...option) error {
	if len(options) == 0 {
		return nil
	}

	err := &xError{
		errs:    nil,
		stack:   nil,
		message: "",
		fields:  nil,
		at:      time.Now().UTC(),
	}
	for _, opt := range options {
		opt(err)
	}
	return err
}

func Unwrap(err error) error {
	return errors.Unwrap(err)
}

func Unwraps(err error) []error {
	if e, ok := err.(interface {
		Unwraps() []error
	}); ok {
		return e.Unwraps()
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

type multiError struct {
	// errs guaranteed to be not empty and squashed (i.e. no nil errors)
	errs []error
}

func (errs multiError) Error() string {
	messages := make([]string, len(errs.errs))
	for i, err := range errs.errs {
		messages[i] = err.Error()
	}
	return strings.Join(messages, "; ")
}

func (errs multiError) Unwrap() []error {
	return errs.errs
}

func Combine(errs ...error) error {
	res := make([]error, 0, len(errs))
	for _, err := range errs {
		if err == nil {
			continue
		}
		res = append(res, err)
	}

	if len(res) > 0 {
		return multiError{res}
	}

	return nil
}

func AppendInto(into *error, errs ...error) {
	if into == nil {
		panic("AppendInto: trying to append into nil")
	}

	multierror, ok := As[multiError](*into)
	if !ok {
		*into = Combine(append(errs, *into)...)
		return
	}

	for _, e := range errs {
		if e == nil {
			continue
		}
		multierror.errs = append(multierror.errs, e)
	}
	*into = multierror
}

func AppendFunc(into *error, f func() error) {
	AppendInto(into, f())
}
