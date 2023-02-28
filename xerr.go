package xerr

import (
	"encoding/json"
	"errors"
	"strings"
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
		res[keyAt] = err.at
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
		// 1 for New function
		xe.callstack = stacktrace(skip + 2)
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

// New - creates error with metadata such as caller information and timestamp.
// Additional metadata can be attached using With* options.
func New(opts ...option) error {
	if len(opts) == 0 {
		return nil
	}

	err := &xError{
		errs:      nil,
		callstack: nil,
		message:   "",
		fields:    map[string]any{},
		at:        time.Now().UTC(),
		caller:    caller(),
		value:     nil,
	}
	for _, opt := range opts {
		opt(err)
	}
	return err
}

// NewM - equivalent to New(WithMessage(message), opts...)
func NewM(message string, opts ...option) error {
	return New(append(opts, WithMessage(message))...)
}

// NewW - equivalent to New(WithErrors(err), opts...)
func NewW(err error, opts ...option) error {
	return New(append(opts, WithErrs(err))...)
}

// NewWM - equivalent to New(WithErrors(err), WithMessage(message), opts...)
func NewWM(err error, message string, opts ...option) error {
	return New(append(opts,
		WithErrs(err),
		WithMessage(message),
	)...)
}

// NewF - equivalent to New(WithMessage(message), WithFields(fields), opts...)
func NewF(message string, fields map[string]any, opts ...option) error {
	return New(append(opts,
		WithMessage(message),
		WithFields(fields),
	)...)
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

// sieveErrs - filter out nil errors
func sieveErrs(errs []error) []error {
	if len(errs) == 0 {
		return nil
	}

	errList := make([]error, 0, len(errs))
	for _, err := range errs {
		if err != nil {
			errList = append(errList, err)
		}
	}
	return errList
}

type multierr struct {
	errs []error
}

func (err multierr) Error() string {
	errsStrings := make([]string, len(err.errs))
	for i, e := range err.errs {
		errsStrings[i] = e.Error()
	}

	return strings.Join(errsStrings, "; ")
}

func (err multierr) Unwraps() []error {
	return err.errs
}

func (err multierr) Unwrap() error {
	return err.errs[0]
}

// Combine multiple errs into single one. If no errors are passed or all of them
// are nil, nil is returned.
func Combine(errs ...error) error {
	if errList := sieveErrs(errs); len(errList) > 0 {
		return multierr{
			errs: errList,
		}
	}

	return nil
}

// Fields - extract fields using `Fields() map[string]any` method. So errors
// constructed with New(WithField/WithFields...) will return those fields.
func Fields(err error) map[string]any {
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

// Unwraps returns the result of calling the Unwraps method on err, if err's
// type contains an Unwraps method returning zero or more errors.
// Otherwise, fallbacks to Unwrap func behavior returning single or none errors.
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

// AppendInto - append errors into `into` error, making it multiple errors error.
// `into` must be not nil.
func AppendInto(into *error, errs ...error) {
	if into == nil {
		panic("AppendInto: trying to append into nil")
	}

	if multierror, ok := As[multierr](*into); ok {
		multierror.errs = append(multierror.errs, sieveErrs(errs)...)
		return
	}

	if multierror, ok := As[*xError](*into); ok {
		multierror.errs = append(multierror.errs, sieveErrs(errs)...)
		return
	}

	*into = Combine(append(errs, *into)...)
}

// AppendFunc - append result of calling f into `into`, `into` must be not nil
func AppendFunc(into *error, f func() error) {
	AppendInto(into, f())
}
