package xerr

import (
	"testing"

	"github.com/rprtr258/assert"
)

func TestCombine_single(t *testing.T) {
	got := Combine(NewM("uuh"))
	_, ok := got.(*Error)
	assert.True(t, ok)
}

func TestCombine_many(t *testing.T) {
	for name, test := range map[string]struct {
		errs    []error
		wantLen int
	}{
		"combine 3": {
			errs:    []error{NewM("1"), NewM("2"), NewM("3")},
			wantLen: 3,
		},
	} {
		t.Run(name, func(t *testing.T) {
			got := Combine(test.errs...).(*multierr)
			assert.Len(t, test.wantLen, got.Errs)
		})
	}
}

func TestCombine_nil(t *testing.T) {
	for name, test := range map[string]struct {
		errs []error
	}{
		"combine nil": {
			errs: nil,
		},
		"combine 0": {
			errs: []error{},
		},
		"combine nils": {
			errs: []error{nil, nil, nil},
		},
	} {
		t.Run(name, func(t *testing.T) {
			got := Combine(test.errs...)
			assert.Zero(t, got)
		})
	}
}
