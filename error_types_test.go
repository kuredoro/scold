package scold_test

import (
	"testing"

	"github.com/kuredoro/scold"
	"github.com/maxatome/go-testdeep/td"
)

func TestStringError(t *testing.T) {
	str1 := "an example of an error"
	err1 := scold.StringError(str1)

	str2 := "an example of an error"
	err2 := scold.StringError(str2)

	str3 := "a different error"
	err3 := scold.StringError(str3)

	td.Cmp(t, err1.Error(), str1)
	td.Cmp(t, err2.Error(), str2)
	td.Cmp(t, err3.Error(), str3)

	if err1 != err2 {
		t.Error("StringErrors of the same string should be equal")
	}

	if err1 == err3 || err2 == err3 {
		t.Errorf("StringErrors of different strings should be not equal")
	}
}
