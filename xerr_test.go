package xerr

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIs(t *testing.T) {
	exampleErr1 := NewM("1")
	exampleErr2 := NewM("2")

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
			err:    NewWM(exampleErr1, "3"),
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

	err1 := NewM("1")
	err2 := NewM("2")
	err3 := NewM("3")

	got, ok := As[multierr](Combine(err1, nil, err2, err3, nil))
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
			err:    NewWM(myErr("inside"), "outside"),
			want:   myErr("inside"),
			wantOk: true,
		},
		"fail": {
			err:    NewWM(NewM("inside"), "outside"),
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

func newErr() error {
	return New(Stacktrace(1))
}

func TestWithStacktrace(tt *testing.T) {
	t := assert.New(tt)

	wantFunctions := []string{
		"github.com/rprtr258/xerr.TestWithStacktrace",
		"testing.tRunner",
		"runtime.goexit",
	}

	err, ok := As[*xError](newErr())
	t.True(ok)

	got := err.callstack
	gotFunctions := make([]string, len(got))
	for i, frame := range got {
		gotFunctions[i] = frame.Function
	}

	t.Equal(wantFunctions, gotFunctions)
}

func TestGetValue(tt *testing.T) {
	t := assert.New(tt)

	err := New(Errors(
		NewM("a"),
		NewM("b", Value(123)),
	))

	intGot, intOk := UnwrapValue[int](err)
	t.True(intOk)
	t.Equal(123, intGot)

	_, boolOk := UnwrapValue[bool](err)
	t.False(boolOk)
}

func TestFields(tt *testing.T) {
	t := assert.New(tt)

	err := New(
		Message("abc"),
		Errors(nil, NewM("def"), nil),
		Value(404),
		Field("field1", 1),
		Fields(map[string]any{
			"field2": "2",
			"field3": 3.3,
		}),
	)
	got := UnwrapFields(err)
	want := map[string]any{
		"field1": 1,
		"field2": "2",
		"field3": 3.3,
	}
	t.Len(got, 8)
	delete(got, "@value")
	delete(got, "@message")
	delete(got, "@caller")
	delete(got, "@errors")
	delete(got, "@at")
	t.Len(got, 3)
	t.Equal(want, got)
}

func faulty() error {
	return New(Message("aboba"))
}

func TestNew_caller(tt *testing.T) {
	t := assert.New(tt)

	err := faulty().(*xError)
	t.Equal("github.com/rprtr258/xerr.faulty", err.caller.Function)
}

func faultyM() error {
	return NewM("aboba")
}

func TestNewM_caller(tt *testing.T) {
	t := assert.New(tt)

	err := faultyM().(*xError)
	t.Equal("github.com/rprtr258/xerr.faultyM", err.caller.Function)
}
