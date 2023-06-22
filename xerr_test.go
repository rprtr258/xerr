package xerr

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCombine(tt *testing.T) {
	t := assert.New(tt)

	err1 := NewM("1")
	err2 := NewM("2")
	err3 := NewM("3")

	got := Combine(err1, nil, err2, err3, nil)
	t.Equal([]error{err1, err2, err3}, got.errs)
}

func newErr() *xError {
	return New(Stacktrace)
}

func TestWithStacktrace(t *testing.T) {
	wantFunctions := []string{
		"github.com/rprtr258/xerr.newErr",
		"github.com/rprtr258/xerr.TestWithStacktrace",
		"testing.tRunner",
		"runtime.goexit",
	}

	err := newErr()

	got := err.callstack
	gotFunctions := make([]string, len(got))
	for i, frame := range got {
		gotFunctions[i] = frame.Function
	}

	assert.Equal(t, wantFunctions, gotFunctions)
}

func TestFields(tt *testing.T) {
	t := assert.New(tt)

	err := New(
		Message("abc"),
		Errors{nil, NewM("def"), nil},
		Fields{"field1": 1},
		Fields{
			"field2": "2",
			"field3": 3.3,
		},
	)
	got := err.toMap()
	want := map[string]any{
		"field1":   1,
		"field2":   "2",
		"field3":   3.3,
		"@message": "abc",
	}
	t.Len(got, 5)
	delete(got, "@errors")
	t.Equal(want, got)
}

func faulty() error {
	Helper()

	return New(Message("aboba"), Caller)
}

func TestNew_caller(tt *testing.T) {
	t := assert.New(tt)

	err := faulty().(*xError)
	t.Equal("github.com/rprtr258/xerr.TestNew_caller", err.caller.Function)
}

func faultyM() error {
	Helper()

	return NewM("aboba", Caller)
}

func TestNewM_caller(tt *testing.T) {
	t := assert.New(tt)

	err := faultyM().(*xError)
	t.Equal("github.com/rprtr258/xerr.TestNewM_caller", err.caller.Function)
}

func TestMarshalJSON(t *testing.T) {
	for name, test := range map[string]struct {
		err  error
		want string
	}{
		"nested multierr": {
			err: Combine(
				errors.New("a"),
				Combine(
					errors.New("b"),
					errors.New("c"),
				),
			),
			want: `["a",["b","c"]]`,
		},
		// TODO: test *xErr
		// "xerr": {
		// 	err: New(
		// 		Message("a"),
		// 		Field("b", 3),
		// 		Errors(errors.New("c")),
		// 	),
		// 	want: `{"@at":"Thu, 16 Mar 2023 01:50:09 UTC","@caller":"/home/rprtr258/pr/xerr/xerr_test.go#github.com/rprtr258/xerr.TestMarshalJSON:200","@errors":["c"],"@message":"a","b":3}`,
		// },
	} {
		t.Run(name, func(t *testing.T) {
			got, err := json.Marshal(test.err)
			assert.NoError(t, err)
			assert.Equal(t, test.want, string(got))
		})
	}
}

func TestJSON(t *testing.T) {
	got, err := json.Marshal(Combine(
		errors.New("a"),
		Combine(
			errors.New("b"),
			errors.New("c"),
		),
	))
	assert.NoError(t, err)
	assert.Equal(t, `["a",["b","c"]]`, string(got))
}

func TestXErr_Error(t *testing.T) {
	got := New(
		Message("aboba"),
		Errors{NewM("123"), NewM("lol")},
		Fields{"code": 404},
		Stacktrace,
	).Error()

	assert.Equal(t,
		strings.Join([]string{
			"aboba",
			"code=404",
			"errs=[123; lol]",
			"stacktrace=[" +
				"/home/rprtr258/pr/xerr/xerr_test.go#github.com/rprtr258/xerr.TestXErr_Error:142; " +
				"/home/rprtr258/.gvm/gos/go1.19.5/src/testing/testing.go#testing.tRunner:1446; " +
				"/home/rprtr258/.gvm/gos/go1.19.5/src/runtime/asm_amd64.s#runtime.goexit:1594; " +
				"]",
		}, " "),
		got,
	)
}
