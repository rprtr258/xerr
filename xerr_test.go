package xerr

import (
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func newErr() *xError {
	return New(Stacktrace).(*xError)
}

func TestWithStacktrace(t *testing.T) {
	wantFunctions := []string{
		"github.com/rprtr258/xerr.newErr",
		"github.com/rprtr258/xerr.TestWithStacktrace",
		"testing.tRunner",
		"runtime.goexit",
	}

	err := newErr()

	got := err.Stacktrace
	gotFunctions := make([]string, len(got))
	for i, frame := range got {
		gotFunctions[i] = frame.Function
	}

	assert.Equal(t, wantFunctions, gotFunctions)
}

func TestFields(t *testing.T) {
	err := New(
		Message("abc"),
		Errors{nil, NewM("def"), nil},
		Fields{"field1": 1},
		Fields{
			"field2": "2",
			"field3": 3.3,
		},
	).(*xError)
	want := map[string]any{
		"field1": 1,
		"field2": "2",
		"field3": 3.3,
	}
	assert.Equal(t, "abc", err.Message)
	assert.Equal(t, want, err.Fields)
	assert.EqualError(t, err.Err, "def")
	assert.Nil(t, err.Errs)
}

func myNewError() error {
	Helper()

	return New(Message("aboba"), Caller)
}

func TestNew_callerHelper(t *testing.T) {
	err := myNewError().(*xError)
	assert.Equal(t, "github.com/rprtr258/xerr.TestNew_callerHelper", err.Caller.Function)
}

func faulty() error {
	return NewM("aboba", Caller)
}

func TestNew_caller(t *testing.T) {
	err := faulty().(*xError)
	assert.Equal(t, "github.com/rprtr258/xerr.faulty", err.Caller.Function)
}

func TestXErr_Error(t *testing.T) {
	got := New(
		Message("aboba"),
		Errors{NewM("123"), NewM("lol")},
		Fields{"code": 404},
		Stacktrace,
	).Error()

	assert.Regexp(t,
		regexp.MustCompile(strings.Join([]string{
			"aboba",
			"code=404",
			regexp.QuoteMeta("errs=[123; lol]"),
			regexp.QuoteMeta("stacktrace=[") +
				".*" + regexp.QuoteMeta("/xerr/xerr_test.go#github.com/rprtr258/xerr.TestXErr_Error:") + `\d+` + "; " +
				".*" + regexp.QuoteMeta("testing.tRunner:") + `\d+` + "; " +
				".*" + regexp.QuoteMeta("runtime.goexit:") + `\d+` + "; " +
				regexp.QuoteMeta("]"),
		}, " ")),
		got,
	)
}
