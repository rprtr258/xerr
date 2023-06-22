package xerr

import (
	"encoding/json"
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

var _ error = (*xError)(nil)

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
	// caller - error origin function frame, nil if no caller added
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
	// addCaller - add caller info
	addCaller bool
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

type callerOpt struct{}

func (o callerOpt) apply(c *xErrorConfig) {
	c.addCaller = true
}

// Caller - add caller
var Caller = callerOpt{}

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

func newx(opts ...Option) *xError {
	Helper()

	if len(opts) == 0 {
		return nil
	}

	config := &xErrorConfig{
		errs:      nil,
		callstack: nil,
		message:   "",
		fields:    map[string]any{},
		at:        time.Now().UTC(),
		addCaller: false,
	}
	for _, opt := range opts {
		opt.apply(config)
	}

	var caller *stackFrame
	if config.addCaller {
		caller = getCaller()
	}

	return &xError{
		errs:      config.errs,
		callstack: config.callstack,
		message:   config.message,
		fields:    config.fields,
		at:        config.at,
		caller:    caller,
	}
}

// New - creates error with metadata such as caller information and timestamp.
// Additional metadata can be attached using With* options.
func New(opts ...Option) *xError {
	Helper()

	return newx(opts...)
}

// NewM - equivalent to New(WithMessage(message), opts...)
func NewM(message Message, opts ...Option) error {
	Helper()

	return newx(append(opts, Message(message))...)
}

// NewW - equivalent to New(WithErrors(err), opts...)
func NewW(err error, opts ...Option) error {
	Helper()

	return newx(append(opts, Errors{err})...)
}

// NewWM - equivalent to New(WithErrors(err), WithMessage(message), opts...)
func NewWM(err error, message Message, opts ...Option) error {
	Helper()

	return newx(append(opts, Errors{err}, Message(message))...)
}

// NewF - equivalent to New(WithMessage(message), WithFields(fields), opts...)
func NewF(message Message, fields map[string]any, opts ...Option) error {
	Helper()

	return newx(append(opts, Fields(fields), Message(message))...)
}

func UnwrapFields(err error) (string, map[string]any) {
	if e, ok := err.(interface {
		UnwrapFields() (string, map[string]any)
	}); ok {
		return e.UnwrapFields()
	}
	return err.Error(), nil
}

// Is - checks target type for type E. Note: that differs from "errors".Is.
// This function does not use Unwrap. To compare errors use ==.
func Is[E error](err error) bool {
	_, ok := err.(E)
	return ok
}

// As - get error as type E. Note: that differs from "errors".As.
// This function does not use Unwrap.
func As[E error](err error) (E, bool) {
	res, ok := err.(E)
	return res, ok
}
