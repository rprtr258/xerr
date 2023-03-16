package xerr

import (
	"bytes"
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

func (err *xError) MarshalJSON() ([]byte, error) {
	return MarshalJSON(err)
}

func (err *xError) toMap() map[string]any {
	res := make(map[string]any, len(err.fields))
	for k, v := range err.fields {
		res[k] = v
	}

	if err.callstack != nil {
		frames := make([]string, len(err.callstack))
		for i, frame := range err.callstack {
			frames[i] = frame.String()
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
		res[keyCaller] = err.caller.String()
	}

	if len(err.errs) != 0 {
		errMessages := make([]any, len(err.errs))
		for i, ierr := range err.errs {
			if xe, ok := ierr.(*xError); ok {
				errMessages[i] = xe.toMap()
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
	bytes, jerr := json.Marshal(err.toMap())
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

type Option func(*xError)

// Errors - wrap errors list, only not nil errors are added
func Errors(errs ...error) Option {
	return func(xe *xError) {
		for _, err := range errs {
			if err != nil {
				xe.errs = append(xe.errs, err)
			}
		}
	}
}

// Stacktrace - add stacktrace
func Stacktrace(skip int) Option {
	return func(xe *xError) {
		// 1 for this callback
		// 1 for New func
		// 1 for newx func
		xe.callstack = stacktrace(skip + 3)
	}
}

// Message - attach error description
func Message(message string) Option {
	return func(xe *xError) {
		xe.message = message
	}
}

// Field - attach single field, old field with same name is overwritten
func Field(name string, value any) Option {
	return func(xe *xError) {
		xe.fields[name] = value
	}
}

// Fields - attach given fields, old fields with such names are overwritten
func Fields(fields map[string]any) Option {
	return func(xe *xError) {
		for name, value := range fields {
			xe.fields[name] = value
		}
	}
}

// Value - attach value to error, if value is nil, no value is attached
func Value(value any) Option {
	return func(xe *xError) {
		xe.value = value
	}
}

func newx(opts ...Option) error {
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
func New(opts ...Option) error {
	return newx(opts...)
}

// NewM - equivalent to New(WithMessage(message), opts...)
func NewM(message string, opts ...Option) error {
	return newx(append(opts, Message(message))...)
}

// NewW - equivalent to New(WithErrors(err), opts...)
func NewW(err error, opts ...Option) error {
	return newx(append(opts, Errors(err))...)
}

// NewWM - equivalent to New(WithErrors(err), WithMessage(message), opts...)
func NewWM(err error, message string, opts ...Option) error {
	return newx(append(opts, Errors(err), Message(message))...)
}

// NewF - equivalent to New(WithMessage(message), WithFields(fields), opts...)
func NewF(message string, fields map[string]any, opts ...Option) error {
	return newx(append(opts, Fields(fields), Message(message))...)
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

// MarshalJSON - marshal error to json
func MarshalJSON(err error) ([]byte, error) {
	switch e := err.(type) {
	case multierr:
		var b bytes.Buffer
		b.WriteRune('[')
		for i, ee := range e.errs {
			bb, errr := MarshalJSON(ee)
			if errr != nil {
				return nil, errr
			}

			if i > 0 {
				b.WriteRune(',')
			}
			b.Write(bb)
		}
		b.WriteRune(']')
		return b.Bytes(), nil
	case *xError:
		return json.Marshal(e.toMap())
	default:
		return json.Marshal(e.Error())
	}
}
