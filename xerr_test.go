package xerr

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIs(t *testing.T) {
	exampleErr1 := New(WithMessage("1"))
	exampleErr2 := New(WithMessage("2"))

	for name, test := range map[string]struct {
		err    error
		target error
		want   bool
	}{
		"same err": {
			err:    exampleErr1,
			target: exampleErr1,
			want:   true,
		},
		"unrelated errs": {
			err:    exampleErr1,
			target: exampleErr2,
			want:   false,
		},
		"wrapped err": {
			err: New(
				WithErr(exampleErr1),
				WithMessage("3"),
			),
			target: exampleErr1,
			want:   true,
		},
	} {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, test.want, Is(test.err, test.target))
		})
	}
}

func TestCombine(tt *testing.T) {
	t := assert.New(tt)

	err1 := New(WithMessage("1"))
	err2 := New(WithMessage("2"))
	err3 := New(WithMessage("3"))

	got, ok := As[*xError](Combine(err1, nil, err2, err3, nil))
	t.True(ok)
	t.Equal([]error{err1, err2, err3}, got.errs)
}

type myErr string

func (err myErr) Error() string {
	return string(err)
}

func TestAs_myErrSuccess(tt *testing.T) {
	t := assert.New(tt)

	want := myErr("inside")
	err := New(
		WithErr(want),
		WithMessage("outside"),
	)

	got, ok := As[myErr](err)
	t.True(ok)
	t.Equal(want, got)
}

func TestAs_myErrFail(tt *testing.T) {
	t := assert.New(tt)

	inside := New(WithMessage("inside"))
	err := New(
		WithErr(inside),
		WithMessage("outside"),
	)

	_, ok := As[myErr](err)
	t.False(ok)
}
