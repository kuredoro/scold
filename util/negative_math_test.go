package util_test

import (
    "testing"
    "fmt"

    "github.com/kuredoro/scold/util"
)

func TestIsPossiblyNegative(t *testing.T) {
    cases := []struct{
        In int
        Want bool
    }{
        {0, false},
        {1, false},
        {-1, true},
        {4294967295, true},
        {4294967294, true},
        {9223372036854775807, false},
    }

    for _, test := range cases {
        t.Run(fmt.Sprint(test.In), func(t *testing.T) {
            got := util.IsPossiblyNegative(test.In)
            if got != test.Want {
                t.Errorf("got %v, want %v", got, test.Want)
            }
        })
    }
}
