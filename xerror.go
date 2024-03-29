package xerr

import (
	"fmt"
	"strings"
	"time"
)

var _ error = (*Error)(nil)

// Error - main structure containing error with metadata
type Error struct {
	// Message describing error
	Message string
	// Fields - error Fields, nil if none
	Fields map[string]any
	// At - error creation timestamp, zero if not specified
	At time.Time

	// errs list of wrapped errors.
	// Either Err or Errs is always nil.
	Err error
	// Errs list of wrapped errors
	// Either Err or Errs is always nil.
	Errs []error

	// Stacktrace of error origin, nil if no Stacktrace added.
	// Either Stacktrace or Caller is always nil.
	Stacktrace []stackFrame
	// Caller - error origin function frame, nil if no Caller added
	// Either Stacktrace or Caller is always nil.
	Caller *stackFrame
}

func (err *Error) Error() string {
	var sb strings.Builder

	if err.Message != "" {
		sb.WriteString(err.Message)
	}

	if !err.At.IsZero() {
		sb.WriteString(" at=")
		sb.WriteString(err.At.Format(time.RFC1123))
	}

	if err.Caller != nil {
		sb.WriteString(" caller=")
		sb.WriteString(err.Caller.String())
	}

	if len(err.Fields) > 0 {
		for k, v := range err.Fields {
			sb.WriteString(" ")
			sb.WriteString(k)
			sb.WriteString("=")
			sb.WriteString(fmt.Sprintf("%+v", v))
		}
	}

	if err.Err != nil {
		sb.WriteString(" err=")
		fmt.Fprintf(&sb, "%q", err.Err.Error())
	}

	if len(err.Errs) != 0 {
		sb.WriteString(" errs=[")
		sb.WriteString(err.Errs[0].Error())
		for _, e := range err.Errs[1:] {
			sb.WriteString("; ")
			sb.WriteString(e.Error())
		}
		sb.WriteString("]")
	}

	if err.Stacktrace != nil {
		sb.WriteString(" stacktrace=[")
		for _, frame := range err.Stacktrace {
			sb.WriteString(frame.String())
			sb.WriteString("; ")
		}
		sb.WriteString("]")
	}

	return sb.String()
}

func (err *Error) Unwrap() error {
	if err.Err != nil {
		return err.Err
	}

	if len(err.Errs) == 0 {
		return nil
	}

	return err.Errs[0]
}

type xErrorConfig struct {
	// errs - list of wrapped errors
	errs []error
	// message describing error
	message string
	// fields - error Fields
	fields map[string]any
	// when - error creation timestamp
	when time.Time
	// addStacktrace which led to xError creation
	addStacktrace bool
	// addCaller function which created xError
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

type stacktraceOpt struct{}

func (o stacktraceOpt) apply(c *xErrorConfig) {
	c.addStacktrace = true
}

// Stacktrace - add stacktrace, overrides Caller
var Stacktrace = stacktraceOpt{}

type callerOpt struct{}

func (o callerOpt) apply(c *xErrorConfig) {
	c.addCaller = true
}

// Caller - add caller
var Caller = callerOpt{}

type whenOpt struct{}

func (o whenOpt) apply(c *xErrorConfig) {
	c.when = time.Now()
}

// When - add timestamp
var When = whenOpt{}

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

func newx(opts ...Option) *Error {
	Helper()

	if len(opts) == 0 {
		return nil
	}

	config := &xErrorConfig{
		errs:          nil,
		addStacktrace: false,
		message:       "",
		fields:        map[string]any{},
		when:          time.Time{},
		addCaller:     false,
	}
	for _, opt := range opts {
		opt.apply(config)
	}

	var callstack []stackFrame
	if config.addStacktrace {
		callstack = stacktrace()
	}

	var caller *stackFrame
	if config.addCaller && !config.addStacktrace {
		caller = getCaller()
	}

	var (
		err  error
		errs []error
	)
	switch len(config.errs) {
	case 0:
	case 1:
		err = config.errs[0]
	default:
		errs = config.errs
	}

	var fields map[string]any
	if len(config.fields) > 0 {
		fields = config.fields
	}

	return &Error{
		Message: config.message,
		Fields:  fields,
		At:      config.when,

		Err:  err,
		Errs: errs,

		Stacktrace: callstack,
		Caller:     caller,
	}
}

// New - creates error with metadata such as caller information and timestamp.
// Additional metadata can be attached using options.
func New(opts ...Option) error {
	Helper()

	return newx(opts...)
}

// NewM - equivalent to New(Message(message), opts...)
func NewM(message string, opts ...Option) error {
	Helper()

	return newx(append(opts, Message(message))...)
}

// NewW - equivalent to New(Errors{err}, opts...)
func NewW(err error, opts ...Option) error {
	Helper()

	return newx(append(opts, Errors{err})...)
}

// NewWM - equivalent to New(Errors{err}, Message(message), opts...)
func NewWM(err error, message string, opts ...Option) *Error {
	Helper()

	return newx(append(opts, Errors{err}, Message(message))...)
}

// NewF - equivalent to New(Message(message), Fields(fields), opts...)
func NewF(message string, fields map[string]any, opts ...Option) error {
	Helper()

	return newx(append(opts, Fields(fields), Message(message))...)
}
