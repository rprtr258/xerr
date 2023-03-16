package xerr

import (
	"strings"
)

type multierr struct {
	errs []error
}

func (err multierr) MarshalJSON() ([]byte, error) {
	return MarshalJSON(err)
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
