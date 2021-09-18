package cptest_test

import (
    "github.com/kuredoro/cptest"
    "testing"
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

    // t.Run("int fields no error", func(t *testing.T) {
    //     type structType struct{
    //         untouched int
    //         i int
    //     }

    //     target := structType{42, 1}

    //     kvm := map[string]string{
    //         "i": "42",
    //     }

    //     want := structType{42, 42}

    //     err := cptest.KVMapUnmarshal(kvm, &target)

    //     td.CmpNoError(t, err)
    //     td.Cmp(t, target, want)
    // })
}
