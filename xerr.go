package xerr

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/davecgh/go-spew/spew"
)

// special metadata keys
var (
	keyStacktrace = "@stacktrace"
	keyValue      = "@value"
	keyMessage    = "@message"
	keyCaller     = "@caller"
	keyErrors     = "@errors"
	keyAt         = "@at"
)

// xError - main structure containing error with metadata
type xError struct {
	// errs list of wrapped errors
	errs []error
	// callstack of error origin, if constructed with WithStack
	callstack []stackFrame
	// message describing error, added using WithMessage
	message string
	// fields added using WithField and WithFields
	fields map[string]any
	// at - error creation timestamp
	at time.Time
	// caller - error origin function frame
	caller *stackFrame
	// value attached to error, nil if none
	// Warning: if nil is attached, value is still nil
	value any
}

func (err *xError) Value() any {
	return err.value
}

func (err *xError) Fields() map[string]any {
	res := make(map[string]any, len(err.fields))
	for k, v := range err.fields {
		res[k] = v
	}

	if err.callstack != nil {
		frames := make([]map[string]any, len(err.callstack))
		for i, frame := range err.callstack {
			frames[i] = map[string]any{
				"function": frame.Function,
				"file":     frame.File,
				"line":     frame.Line,
			}
		}
		res[keyStacktrace] = frames
	}

	if err.value != nil {
		res[keyValue] = err.value
	}

	if err.message != "" {
		res[keyMessage] = err.message
	}

	if err.caller != nil {
		res[keyCaller] = err.caller
	}

	if len(err.errs) != 0 {
		errMessages := make([]any, len(err.errs))
		for i, ierr := range err.errs {
			if xe, ok := ierr.(*xError); ok {
				errMessages[i] = xe.Fields()
			} else {
				errMessages[i] = ierr.Error()
			}
		}
		res[keyErrors] = errMessages
	}

	if !err.at.IsZero() {
		res[keyAt] = err.at.Format(time.RFC1123)
	}

	return res
}

func (err *xError) Error() string {
	bytes, jerr := json.Marshal(err.Fields())
	if jerr != nil {
		return spew.Sprint(err)
	}

	return string(bytes)
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

type option func(*xError)

// WithErrs - wrap errors list, only not nil errors are added
func WithErrs(errs ...error) option {
	return func(xe *xError) {
		for _, err := range errs {
			if err != nil {
				xe.errs = append(xe.errs, err)
			}
		}
	}
}

// WithStack - add stacktrace
func WithStack(skip int) option {
	return func(xe *xError) {
		// 1 for this callback
		// 1 for New func
		// 1 for newx func
		xe.callstack = stacktrace(skip + 3)
	}
}

// WithMessage - attach error description
func WithMessage(message string) option {
	return func(xe *xError) {
		xe.message = message
	}
}

// WithField - attach single field, old field with same name is overwritten
func WithField(name string, value any) option {
	return func(xe *xError) {
		xe.fields[name] = value
	}
}

// WithFields - attach given fields, old fields with such names are overwritten
func WithFields(fields map[string]any) option {
	return func(xe *xError) {
		for name, value := range fields {
			xe.fields[name] = value
		}
	}
}

// WithValue - attach value to error, if value is nil, no value is attached
func WithValue(value any) option {
	return func(xe *xError) {
		xe.value = value
	}
}

func newx(opts ...option) error {
	if len(opts) == 0 {
		return nil
	}

	err := &xError{
		errs:      nil,
		callstack: nil,
		message:   "",
		fields:    map[string]any{},
		at:        time.Now().UTC(),
		caller:    caller(1),
		value:     nil,
	}
	for _, opt := range opts {
		opt(err)
	}
	return err
}

// New - creates error with metadata such as caller information and timestamp.
// Additional metadata can be attached using With* options.
func New(opts ...option) error {
	return newx(opts...)
}

// NewM - equivalent to New(WithMessage(message), opts...)
func NewM(message string, opts ...option) error {
	return newx(append(opts, WithMessage(message))...)
}

// NewW - equivalent to New(WithErrors(err), opts...)
func NewW(err error, opts ...option) error {
	return newx(append(opts, WithErrs(err))...)
}

// NewWM - equivalent to New(WithErrors(err), WithMessage(message), opts...)
func NewWM(err error, message string, opts ...option) error {
	return newx(append(opts, WithErrs(err), WithMessage(message))...)
}

// NewF - equivalent to New(WithMessage(message), WithFields(fields), opts...)
func NewF(message string, fields map[string]any, opts ...option) error {
	return newx(append(opts, WithFields(fields), WithMessage(message))...)
}

// UnwrapValue from err having type T and bool detecting if such value was found.
// First value found in depth-first order is returned. Value is extracted using
// `Value() any` method, so errors constructed with New(WithValue(value)) will
// give value out.
func UnwrapValue[T any](err error) (T, bool) {
	stack := []error{err}
	for len(stack) > 0 {
		cur := stack[0]
		stack = stack[1:]
		if e, ok := cur.(interface {
			Value() any
		}); ok {
			if value, ok2 := e.Value().(T); ok2 {
				return value, true
			}
		}
		stack = append(stack, Unwraps(cur)...)
	}

	var zero T
	return zero, false
}

// UnwrapFields using `UnwrapFields() map[string]any` method. So errors
// constructed with New(WithField/WithFields...) will return those fields.
func UnwrapFields(err error) map[string]any {
	if e, ok := err.(interface {
		Fields() map[string]any
	}); ok {
		return e.Fields()
	}

	return nil
}

// Unwrap is alias to "errors".Unwrap
func Unwrap(err error) error {
	return errors.Unwrap(err)
}

// Is - alias for "errors".Is func
func Is(err, target error) bool {
	return errors.Is(err, target)
}

// As - generic alias to "errors".As func
func As[E error](err error) (E, bool) {
	var res E
	ok := errors.As(err, &res)
	return res, ok
}
