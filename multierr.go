package xerr

import (
	"strconv"
	"strings"
)

var _ error = (*multierr)(nil)

type multierr struct {
	Errs []error
}

func (err multierr) Error() string {
	errsStrings := make([]string, len(err.Errs))
	for i, e := range err.Errs {
		errsStrings[i] = e.Error()
	}

	return strings.Join(errsStrings, "; ")
}

func (err multierr) Unwrap() error {
	return err.Errs[0]
}

func (err multierr) UnwrapFields() (string, map[string]any) {
	fields := make(map[string]any, len(err.Errs))
	for i, e := range err.Errs {
		fields[strconv.Itoa(i)] = e
	}
	return "", fields
}

// Combine multiple errs into single one. If no errors are passed or all of them
// are nil, nil is returned.
func Combine(errs ...error) *multierr {
	if errList := appendErrs(nil, errs); len(errList) > 0 {
		return &multierr{
			Errs: errList,
		}
	}

	return nil
}

// AppendInto - append errors into `into` error, making it multiple errors error.
// `into` must be not nil.
func AppendInto(into *error, errs ...error) {
	if into == nil {
		panic("AppendInto: trying to append into nil")
	}

	if *into == nil {
		if len(errs) == 1 {
			*into = errs[0]
		} else {
			*into = Combine(errs...)
		}
		return
	}

	switch err := (*into).(type) {
	case *multierr:
		err.Errs = append(err.Errs, errs...)
	case *xError:
		err.Errs = appendErrs(err.Errs, errs)
	default:
		*into = Combine(append(errs, *into)...)
	}
}

// AppendFunc - append result of calling f into `into`, `into` must be not nil
func AppendFunc(into *error, f func() error) {
	AppendInto(into, f())
}

// appendErrs - filter out nil errors
func appendErrs(into []error, errs []error) []error {
	cnt := 0
	for _, err := range errs {
		if err != nil {
			cnt++
		}
	}

	res := into
	if cap(res)-len(res) < cnt {
		res = make([]error, 0, len(into)+cnt)
		copy(res, into)
	}
	for _, err := range errs {
		if err != nil {
			res = append(res, err)
		}
	}
	return res
}
