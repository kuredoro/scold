package cptest_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/go-multierror"
	"github.com/kuredoro/cptest"
	"github.com/maxatome/go-testdeep/td"
)

func TestKVMapUnmarshal(t *testing.T) {
    t.Run("empty map and struct", func(t *testing.T) {
        var target struct{}

        kvm := map[string]string{}

        err := cptest.KVMapUnmarshal(kvm, &target)

        td.CmpNoError(t, err)
    })

    t.Run("empty map doesn't affect structs", func(t *testing.T) {
        target := struct{
            foo int
            bar float64
            str string
            m map[string]int
        }{
            42, 11.0, "struct", make(map[string]int),
        }

        kvm := map[string]string{}

        want := target

        err := cptest.KVMapUnmarshal(kvm, &target)

        td.CmpNoError(t, err)
        td.Cmp(t, target, want)
    })

    t.Run("unmarshal only works on structs or pointers to them", func(t *testing.T) {
        kvm := map[string]string{}

        err := cptest.KVMapUnmarshal(kvm, 42)
        td.CmpError(t, err, cptest.NotAStructLike)

        i := 42
        err = cptest.KVMapUnmarshal(kvm, &i)
        td.CmpError(t, err, cptest.NotAStructLike)

        err = cptest.KVMapUnmarshal(kvm, "foo")
        td.CmpError(t, err, cptest.NotAStructLike)

        str := "foo"
        err = cptest.KVMapUnmarshal(kvm, &str)
        td.CmpError(t, err, cptest.NotAStructLike)

        err = cptest.KVMapUnmarshal(kvm, []int{1, 2, 3})
        td.CmpError(t, err, cptest.NotAStructLike)

        err = cptest.KVMapUnmarshal(kvm, [...]int{1, 2, 3})
        td.CmpError(t, err, cptest.NotAStructLike)

        err = cptest.KVMapUnmarshal(kvm, kvm)
        td.CmpError(t, err, cptest.NotAStructLike)

        err = cptest.KVMapUnmarshal(kvm, func() {})
        td.CmpError(t, err, cptest.NotAStructLike)

        err = cptest.KVMapUnmarshal(kvm, make(chan int))
        td.CmpError(t, err, cptest.NotAStructLike)

        // ---

        err = cptest.KVMapUnmarshal(kvm, struct{}{})
        td.CmpNoError(t, err)

        test := struct{}{}
        err = cptest.KVMapUnmarshal(kvm, &test)
        td.CmpNoError(t, err)
    })

    t.Run("report missing fields", func(t *testing.T) {
        target := struct{}{}

        kvm := map[string]string{
            "Foo": "42",
            "Bar": "ハロー",
            "AGAIN?": "435",
        }

        errs := cptest.KVMapUnmarshal(kvm, &target).(*multierror.Error)

        td.CmpError(t, errs)

        wantErrors := []error{
            &cptest.MissingFieldError{"Foo"},
            &cptest.MissingFieldError{"Bar"},
            &cptest.MissingFieldError{"AGAIN?"},
        }

        td.Cmp(t, errs.Errors, wantErrors)
    })

    t.Run("int fields, no error", func(t *testing.T) {
        type structType struct{
            Untouched int
            I int
            Ui uint
            I8 int8
            I16 int16
            I32 int32
            I64 int64
            U8 uint8
            U16 uint16
            U32 uint32
            U64 uint64
        }

        target := structType{42, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

        kvm := map[string]string{
            "I": "42",
            "I8": "127",
            "I16": "-32000",
            "I32": "-2000000000",
            "I64": "18000000000",
            "Ui": "12",
            "U8": "255",
            "U16": "65000",
            "U32": "4000000000",
            "U64": "32000000000",
        }

        want := structType{42, 42, 12, 127, -32000, -2000000000, 18000000000, 255, 65000, 4000000000, 32000000000}

        err := cptest.KVMapUnmarshal(kvm, &target)

        td.CmpNoError(t, err)
        td.Cmp(t, target, want)
    })
}
