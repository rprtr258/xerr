package xerr

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIs(t *testing.T) {
	exampleErr1 := New(WithMessage("1"))
	exampleErr2 := New(WithMessage("2"))

	for _, test := range []struct {
		name   string
		err    error
		target error
		want   bool
	}{
		{
			name:   "same err",
			err:    exampleErr1,
			target: exampleErr1,
			want:   true,
		},
		{
			name:   "unrelated errs",
			err:    exampleErr1,
			target: exampleErr2,
			want:   false,
		},
		{
			name: "wrapped err",
			err: New(
				WithErr(exampleErr1),
				WithMessage("3"),
			),
			target: exampleErr1,
			want:   true,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			got := Is(test.err, test.target)
			assert.Equal(t, test.want, got)
		})
	}
}

func TestCombine(t *testing.T) {
	err1 := New(WithMessage("1"), WithNoTimestamp())
	err2 := New(WithMessage("2"), WithNoTimestamp())
	err3 := New(WithMessage("3"), WithNoTimestamp())

	got := Combine(err1, nil, err2, err3, nil)

	want := `{"errors":[` +
		`{"message":"1"},` +
		`{"message":"2"},` +
		`{"message":"3"}` +
		`]}`
	assert.Equal(t, want, got.Error())
}

type myErr string

func (err myErr) Error() string {
	return string(err)
}

func TestAs_myErrSuccess(t *testing.T) {
	want := myErr("inside")
	err := New(
		WithErr(want),
		WithMessage("outside"),
	)

	got, ok := As[myErr](err)
	assert.True(t, ok)
	assert.Equal(t, want, got)
}

func TestAs_myErrFail(t *testing.T) {
	inside := New(WithMessage("inside"))
	err := New(
		WithErr(inside),
		WithMessage("outside"),
	)

	_, ok := As[myErr](err)
	assert.False(t, ok)
}
