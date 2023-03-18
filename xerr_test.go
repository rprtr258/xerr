package xerr

import (
	"encoding/json"
	"errors"
	"strings"
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
	got := err.(*xError).toMap()
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

func TestMarshalJSON(t *testing.T) {
	for name, test := range map[string]struct {
		err  error
		want string
	}{
		"foreign error": {
			err:  errors.New("a"),
			want: `"a"`,
		},
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
		// 		Value(404),
		// 	),
		// 	want: `{"@at":"Thu, 16 Mar 2023 01:50:09 UTC","@caller":"/home/rprtr258/pr/xerr/xerr_test.go#github.com/rprtr258/xerr.TestMarshalJSON:200","@errors":["c"],"@message":"a","@value":404,"b":3}`,
		// },
	} {
		t.Run(name, func(t *testing.T) {
			got, err := MarshalJSON(test.err)
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

func concat(parts ...[]string) []string {
	res := []string(nil)
	for _, part := range parts {
		res = append(res, part...)
	}
	return res
}

func TestXErr_Error(t *testing.T) {
	err := New(
		Message("aboba"),
		Errors(NewM("123"), NewM("lol")),
		Field("code", 404),
	)
	rawWords := strings.Split(err.Error(), " ")
	words := concat(rawWords[:1], rawWords[6:10], rawWords[15:18], rawWords[23:])

	assert.Equal(t,
		[]string{
			/*  0 */ "aboba",
			/*  1 */ // "at=Sat,",
			/*  2 */ // "18",
			/*  3 */ // "Mar",
			/*  4 */ // "2023",
			/*  5 */ // "10:01:40",
			/*  6 */ "UTC",
			/*  7 */ "caller=/home/rprtr258/pr/xerr/xerr_test.go#github.com/rprtr258/xerr.TestXErr_Error:241",
			/*  8 */ "code=404",
			/*  9 */ "errs=[123",
			/* 10 */ // "at=Sat,",
			/* 11 */ // "18",
			/* 12 */ // "Mar",
			/* 13 */ // "2023",
			/* 14 */ // "10:01:40",
			/* 15 */ "UTC",
			/* 16 */ "caller=/home/rprtr258/pr/xerr/xerr_test.go#github.com/rprtr258/xerr.TestXErr_Error:243;",
			/* 17 */ "lol",
			/* 18 */ // "at=Sat,",
			/* 19 */ // "18",
			/* 20 */ // "Mar",
			/* 21 */ // "2023",
			/* 22 */ // "10:01:40",
			/* 23 */ "UTC",
			/* 24 */ "caller=/home/rprtr258/pr/xerr/xerr_test.go#github.com/rprtr258/xerr.TestXErr_Error:243]",
		},
		words,
	)
}
