package cptest_test

import (
	"errors"
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
        target := struct{Exists int}{42}

        kvm := map[string]string{
            "Foo": "42",
            "Bar": "ハロー",
            "AGAIN?": "435",
            "Exists": "Whaa?",
        }

        errs := cptest.KVMapUnmarshal(kvm, &target).(*multierror.Error)

        td.CmpError(t, errs)

        wantMissingFields := map[string]struct{}{
            "Foo": {},
            "Bar": {},
            "AGAIN?": {},
        }

        gotMissingFields := map[string]struct{}{}
        for _, err := range errs.Errors {
            var fieldError *cptest.MissingFieldError
            if !errors.As(err, &fieldError) {
                t.Errorf("error list contains an error of type different than MissingFieldError (%#v)", err)
                continue
            }

            missingField := fieldError.FieldName

            _, inInput := kvm[missingField]
            if !inInput {
                t.Errorf("error list contains an error for a missing field %q that wasn't specified in the input map", missingField)
                continue
            }

            _, seenBefore := gotMissingFields[missingField]
            if seenBefore {
                t.Errorf("error list contains a duplicate of an error for a missing field %q", missingField)
                continue
            }

            gotMissingFields[missingField] = struct{}{}
        }

        if len(gotMissingFields) != len(wantMissingFields) {
            t.Errorf("Some missing fields weren't detected: got %v, want %v", gotMissingFields, wantMissingFields)
        }
    })

    // t.Run("int fields no error", func(t *testing.T) {
    //     type structType struct{
    //         Untouched int
    //         I int
    //     }

    //     target := structType{42, 1}

    //     kvm := map[string]string{
    //         "I": "42",
    //     }

    //     want := structType{42, 42}

    //     err := cptest.KVMapUnmarshal(kvm, &target)

    //     td.CmpNoError(t, err)
    //     td.Cmp(t, target, want)
    // })
}
