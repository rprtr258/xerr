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
				WithErrs(exampleErr1),
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

func TestAs(tt *testing.T) {
	for name, test := range map[string]struct {
		err    error
		want   error
		wantOk bool
	}{
		"success": {
			err: New(
				WithErrs(myErr("inside")),
				WithMessage("outside"),
			),
			want:   myErr("inside"),
			wantOk: true,
		},
		"fail": {
			err: New(
				WithErrs(New(WithMessage("inside"))),
				WithMessage("outside"),
			),
			want:   nil,
			wantOk: false,
		},
	} {
		tt.Run(name, func(tt *testing.T) {
			t := assert.New(tt)

			got, ok := As[myErr](test.err)
			t.Equal(test.wantOk, ok)
			if ok {
				t.Equal(test.want, got)
			}
		})
	}
}

func newErr() *xError {
	return New(WithStack(1))
}

func TestWithStacktrace(tt *testing.T) {
	t := assert.New(tt)

	wantFunctions := []string{
		"github.com/rprtr258/xerr.TestWithStacktrace",
		"testing.tRunner",
		"runtime.goexit",
	}

	got := newErr().stack
	gotFunctions := make([]string, len(got))
	for i, frame := range got {
		gotFunctions[i] = frame.Function
	}

	t.Equal(wantFunctions, gotFunctions)
}

func TestGetValue(t *testing.T) {
	err := New(WithErrs(
		New(WithMessage("a")),
		New(WithMessage("b"), WithValue(123)),
	))

	intGot, intOk := GetValue[int](err)
	assert.True(t, intOk)
	assert.Equal(t, 123, intGot)

	_, boolOk := GetValue[bool](err)
	assert.False(t, boolOk)
}
