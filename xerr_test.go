package xerr_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/rprtr258/xerr"
)

func TestIs(t *testing.T) {
	exampleErr1 := xerr.New("1")
	exampleErr2 := xerr.New("2")

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
			name:   "wrapped err",
			err:    xerr.Wrap(exampleErr1, "3"),
			target: exampleErr1,
			want:   true,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			got := xerr.Is(test.err, test.target)
			assert.Equal(t, test.want, got)
		})
	}
}

type myErr string

func (err myErr) Error() string {
	return string(err)
}

func TestAs_myErrSuccess(t *testing.T) {
	want := myErr("inside")
	err := xerr.Wrap(want, "outside")

	got, ok := xerr.As[myErr](err)
	assert.True(t, ok)
	assert.Equal(t, want, got)
}

func TestAs_myErrFail(t *testing.T) {
	inside := xerr.New("inside")
	err := xerr.Wrap(inside, "outside")

	_, ok := xerr.As[myErr](err)
	assert.False(t, ok)
}
