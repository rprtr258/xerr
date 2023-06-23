package xerr

// Is - checks target type for type E. Note: that differs from "errors.Is".
// This function does not use Unwrap.
func Is[E error](err error) bool {
	_, ok := err.(E)
	return ok
}

// As - get error as type E. Note: that differs from "errors.As".
// This function does not use Unwrap.
func As[E error](err error) (E, bool) {
	res, ok := err.(E)
	return res, ok
}
