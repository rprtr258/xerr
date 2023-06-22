package xerr

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

// special metadata keys
var (
	keyStacktrace = "@stacktrace"
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
}

func (err *xError) MarshalJSON() ([]byte, error) {
	return json.Marshal(err.toMap())
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

	if err.message != "" {
		res[keyMessage] = err.message
	}

	if err.caller != nil {
		res[keyCaller] = err.caller.String()
	}

	if len(err.errs) != 0 {
		res[keyErrors] = err.errs
	}

	if !err.at.IsZero() {
		res[keyAt] = err.at.Format(time.RFC1123)
	}

	return res
}

func (err *xError) Error() string {
	var sb strings.Builder

	if err.message != "" {
		sb.WriteString(err.message)
	}

	if !err.at.IsZero() {
		sb.WriteString(" at=")
		sb.WriteString(err.at.Format(time.RFC1123))
	}

	if err.caller != nil {
		sb.WriteString(" caller=")
		sb.WriteString(err.caller.String())
	}

	if len(err.fields) > 0 {
		for k, v := range err.fields {
			sb.WriteString(" ")
			sb.WriteString(k)
			sb.WriteString("=")
			sb.WriteString(fmt.Sprintf("%+v", v))
		}
	}

	if len(err.errs) != 0 {
		sb.WriteString(" errs=[")
		sb.WriteString(err.errs[0].Error())
		for _, e := range err.errs[1:] {
			sb.WriteString("; ")
			sb.WriteString(e.Error())
		}
		sb.WriteString("]")
	}

	if err.callstack != nil {
		sb.WriteString(" stacktrace=[")
		for _, frame := range err.callstack {
			sb.WriteString(frame.String())
			sb.WriteString("; ")
		}
		sb.WriteString("]")
	}

	return sb.String()
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

func (err *xError) UnwrapFields() (string, map[string]any) {
	// TODO: simplify
	fields := err.toMap()
	delete(fields, keyMessage)
	return err.message, fields
}

// TODO: simplify
type xErrorConfig struct {
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
	// callerSkip - how many skips before getting caller
	callerSkip int
}

type Option interface {
	apply(*xErrorConfig)
}

// Errors - wrap errors list, only not nil errors are added
type Errors []error

func (o Errors) apply(c *xErrorConfig) {
	for _, err := range o {
		if err != nil {
			c.errs = append(c.errs, err)
		}
	}
}

type stacktraceOpt struct {
	skip int
}

func (o stacktraceOpt) apply(c *xErrorConfig) {
	// 1 for this callback
	// 1 for New func
	// 1 for newx func
	c.callstack = stacktrace(o.skip + 3)
}

// Stacktrace - add stacktrace
func Stacktrace(skip int) Option {
	return stacktraceOpt{
		skip: skip,
	}
}

type callerSkipOpt struct {
	skip int
}

func (o callerSkipOpt) apply(c *xErrorConfig) {
	c.callerSkip += o.skip
}

// CallerSkip - add caller skip
func CallerSkip(skip int) Option {
	return callerSkipOpt{
		skip: skip,
	}
}

type Message string

// Message - attach error description
func (o Message) apply(c *xErrorConfig) {
	c.message = string(o)
}

// Fields - attach given fields, old fields with such names are overwritten
type Fields map[string]any

func (o Fields) apply(c *xErrorConfig) {
	for name, value := range o {
		c.fields[name] = value
	}
}

func newx(opts ...Option) error {
	if len(opts) == 0 {
		return nil
	}

	config := &xErrorConfig{
		errs:       nil,
		callstack:  nil,
		message:    "",
		fields:     map[string]any{},
		at:         time.Now().UTC(),
		callerSkip: 1,
	}
	for _, opt := range opts {
		opt.apply(config)
	}

	return &xError{
		errs:      config.errs,
		callstack: config.callstack,
		message:   config.message,
		fields:    config.fields,
		at:        config.at,
		caller:    caller(config.callerSkip),
	}
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
	return newx(append(opts, Errors{err})...)
}

// NewWM - equivalent to New(WithErrors(err), WithMessage(message), opts...)
func NewWM(err error, message string, opts ...Option) error {
	return newx(append(opts, Errors{err}, Message(message))...)
}

// NewF - equivalent to New(WithMessage(message), WithFields(fields), opts...)
func NewF(message string, fields map[string]any, opts ...Option) error {
	return newx(append(opts, Fields(fields), Message(message))...)
}

// Unwrap is alias to "errors".Unwrap
func Unwrap(err error) error {
	return errors.Unwrap(err)
}

func UnwrapFields(err error) (string, map[string]any) {
	if e, ok := err.(interface {
		UnwrapFields() (string, map[string]any)
	}); ok {
		return e.UnwrapFields()
	}
	return err.Error(), nil
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
