package xerr

import (
	"strings"
)

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
