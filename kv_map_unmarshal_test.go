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
