package cptest

import (
	"reflect"
	"testing"

	"github.com/sanity-io/litter"
)

func AssertTests(t *testing.T, got Inputs, want []Test) {
    t.Helper()
    if !reflect.DeepEqual(got.Tests, want) {
        t.Errorf("\ngot %v\nwant %v", litter.Sdump(got.Tests), litter.Sdump(want))
    }
}
