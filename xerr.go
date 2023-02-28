package xerr

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

type xError struct {
	// errs guranteed to be not nil
	errs    []error
	stack   []stackFrame
	message string
	fields  map[string]any
	at      time.Time
	caller  *stackFrame
	value   any
}

func (err *xError) Fields() map[string]any {
	return err.fields
}

func (err *xError) fill() {
	if err.stack != nil {
		frames := make([]map[string]any, len(err.stack))
		for i, frame := range err.stack {
			frames[i] = map[string]any{
				"function": frame.Function,
				"file":     frame.File,
				"line":     frame.Line,
			}
		}
		err.fields["@stacktrace"] = frames
	}

	if err.value != nil {
		err.fields["@value"] = err.value
	}

	if err.message != "" {
		err.fields["@message"] = err.message
	}

	err.fields["@caller"] = err.caller

	if len(err.errs) != 0 {
		errMessages := make([]any, len(err.errs))
		for i, ierr := range err.errs {
			if xe, ok := ierr.(*xError); ok {
				xe.fill()
				errMessages[i] = xe.fields
			} else {
				errMessages[i] = ierr.Error()
			}
		}
		err.fields["@errors"] = errMessages
	}

	if !err.at.IsZero() {
		err.fields["@at"] = err.at.Format(time.RFC3339)
	}
}

func (err *xError) Error() string {
	err.fill()

	bytes, jerr := json.Marshal(err.fields)
	if jerr != nil {
		return fmt.Sprintf("%#v", err.fields)
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

func WithErrs(errs ...error) option {
	return func(xe *xError) {
		for _, err := range errs {
			if err != nil {
				xe.errs = append(xe.errs, err)
			}
		}
	}
}

func WithStack(skip int) option {
	return func(xe *xError) {
		// 1 for this callback
		// 1 for New function
		xe.stack = stacktrace(skip + 2)
	}
}

func WithMessage(message string) option {
	return func(xe *xError) {
		xe.message = message
	}
}

func WithField(name string, value any) option {
	return func(xe *xError) {
		xe.fields[name] = value
	}
}

func WithFields(fields map[string]any) option {
	return func(xe *xError) {
		for name, value := range fields {
			xe.fields[name] = value
		}
	}
}

func WithValue(value any) option {
	return func(xe *xError) {
		xe.value = value
	}
}

func New(opts ...option) *xError {
	if len(opts) == 0 {
		return nil
	}

	err := &xError{
		errs:    nil,
		stack:   nil,
		message: "",
		fields:  map[string]any{},
		at:      time.Now().UTC(),
		caller:  caller(),
		value:   nil,
	}
	for _, opt := range opts {
		opt(err)
	}
	return err
}

func NewM(message string, opts ...option) *xError {
	return New(append(opts, WithMessage(message))...)
}

func NewW(err error, opts ...option) *xError {
	return New(append(opts, WithErrs(err))...)
}

func NewWM(err error, message string, opts ...option) *xError {
	return New(append(opts,
		WithErrs(err),
		WithMessage(message),
	)...)
}

func NewF(message string, fields map[string]any, opts ...option) *xError {
	return New(append(opts,
		WithMessage(message),
		WithFields(fields),
	)...)
}

func GetValue[T any](err error) (T, bool) {
	stack := []error{err}
	for len(stack) > 0 {
		cur := stack[0]
		stack = stack[1:]
		if e, ok := cur.(*xError); ok {
			if value, ok2 := e.value.(T); ok2 {
				return value, true
			}
		}
		stack = append(stack, Unwraps(cur)...)
	}

	var zero T
	return zero, false
}

func sieveErrs(errs []error) []error {
	errList := make([]error, 0, len(errs))
	for _, err := range errs {
		if err != nil {
			errList = append(errList, err)
		}
	}
	return errList
}

func Combine(errs ...error) error {
	if len(errs) == 0 {
		return nil
	}

	if errList := sieveErrs(errs); len(errList) > 0 {
		return &xError{
			errs:    errList,
			stack:   nil,
			message: "",
			fields:  nil,
			at:      time.Time{},
			caller:  nil,
			value:   nil,
		}
	}

	return nil
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
	if errs, ok := As[*xError](err); ok {
		return errs.errs
	}

	if err != nil {
		return []error{err}
	}

	return nil
}

func AppendInto(into *error, errs ...error) {
	if into == nil {
		panic("AppendInto: trying to append into nil")
	}

	multierror, ok := As[*xError](*into)
	if !ok {
		*into = Combine(append(errs, *into)...)
		return
	}

	multierror.errs = append(multierror.errs, sieveErrs(errs)...)
}

func AppendFunc(into *error, f func() error) {
	AppendInto(into, f())
}
