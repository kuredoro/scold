package cptest_test

import (
	"reflect"
	"testing"

	"github.com/hashicorp/go-multierror"
	"github.com/kuredoro/cptest"
	"github.com/maxatome/go-testdeep/td"
)

func init() {
	td.DefaultContextConfig.UseEqual = true
	td.DefaultContextConfig.MaxErrors = -1
}

func TestKVMapUnmarshal(t *testing.T) {
	t.Run("empty map and struct", func(t *testing.T) {
		var target struct{}

		kvm := map[string]string{}

		err := cptest.KVMapUnmarshal(kvm, &target)

		td.CmpNoError(t, err)
	})

	t.Run("empty map doesn't affect structs", func(t *testing.T) {
		target := struct {
			foo int
			bar float64
			str string
			m   map[string]int
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

		td.CmpPanic(t, func() { cptest.KVMapUnmarshal(kvm, 42) }, cptest.NotAStructLike)

		i := 42
		td.CmpPanic(t, func() { cptest.KVMapUnmarshal(kvm, &i) }, cptest.NotAStructLike)

		td.CmpPanic(t, func() { cptest.KVMapUnmarshal(kvm, "foo") }, cptest.NotAStructLike)

		str := "foo"
		td.CmpPanic(t, func() { cptest.KVMapUnmarshal(kvm, &str) }, cptest.NotAStructLike)

		td.CmpPanic(t, func() { cptest.KVMapUnmarshal(kvm, []int{1, 2, 3}) }, cptest.NotAStructLike)

		td.CmpPanic(t, func() { cptest.KVMapUnmarshal(kvm, [...]int{1, 2, 3}) }, cptest.NotAStructLike)

		td.CmpPanic(t, func() { cptest.KVMapUnmarshal(kvm, kvm) }, cptest.NotAStructLike)

		td.CmpPanic(t, func() { cptest.KVMapUnmarshal(kvm, func() {}) }, cptest.NotAStructLike)

		td.CmpPanic(t, func() { cptest.KVMapUnmarshal(kvm, make(chan int)) }, cptest.NotAStructLike)

		// ---

		err := cptest.KVMapUnmarshal(kvm, struct{}{})
		td.CmpNoError(t, err)

		test := struct{}{}
		err = cptest.KVMapUnmarshal(kvm, &test)
		td.CmpNoError(t, err)
	})

	t.Run("report missing fields", func(t *testing.T) {
		target := struct{}{}

		kvm := map[string]string{
			"Foo":    "42",
			"Bar":    "ハロー",
			"AGAIN?": "435",
		}

		errs := cptest.KVMapUnmarshal(kvm, &target).(*multierror.Error)

		td.CmpError(t, errs)

		wantErrs := []error{
			&cptest.MissingFieldError{"Foo"},
			&cptest.MissingFieldError{"Bar"},
			&cptest.MissingFieldError{"AGAIN?"},
		}

		td.Cmp(t, errs.Errors, td.Bag(td.Flatten(wantErrs)))
	})

	t.Run("int fields, no error", func(t *testing.T) {
		type structType struct {
			Untouched int
			I         int
			Ui        uint
			I8        int8
			I16       int16
			I32       int32
			I64       int64
			U8        uint8
			U16       uint16
			U32       uint32
			U64       uint64
		}

		target := structType{42, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

		kvm := map[string]string{
			"I":   "42",
			"I8":  "127",
			"I16": "-32000",
			"I32": "-2000000000",
			"I64": "18000000000",
			"Ui":  "12",
			"U8":  "255",
			"U16": "65000",
			"U32": "4000000000",
			"U64": "32000000000",
		}

		want := structType{42, 42, 12, 127, -32000, -2000000000, 18000000000, 255, 65000, 4000000000, 32000000000}

		err := cptest.KVMapUnmarshal(kvm, &target)

		td.CmpNoError(t, err)
		td.Cmp(t, target, want)
	})

	t.Run("int fields, values out of bounds or bogus", func(t *testing.T) {
		type structType struct {
			Untouched int
			I         int
			Ui        uint
			I8        int8
			I16       int16
			I32       int32
			I64       int64
			U8        uint8
			U16       uint16
			U32       uint32
			U64       uint64
		}

		target := structType{42, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

		kvm := map[string]string{
			"I":   "-0",
			"I8":  "-129",
			"I16": "40000",
			"I32": "-3000000000",
			"I64": "10000000000000000000",
			"Ui":  "０",
			"U8":  "300",
			"U16": "67000",
			"U32": "5000000000",
			"U64": "20000000000000000000",
		}

		want := structType{42, 0, 2, 3, 4, 5, 6, 7, 8, 9, 10}

		errs := cptest.KVMapUnmarshal(kvm, &target).(*multierror.Error)

		wantErrs := []error{
			&cptest.NotValueOfType{reflect.Int8, "-129"},
			&cptest.NotValueOfType{reflect.Int16, "40000"},
			&cptest.NotValueOfType{reflect.Int32, "-3000000000"},
			&cptest.NotValueOfType{reflect.Int64, "10000000000000000000"},
			&cptest.NotValueOfType{reflect.Uint, "０"},
			&cptest.NotValueOfType{reflect.Uint8, "300"},
			&cptest.NotValueOfType{reflect.Uint16, "67000"},
			&cptest.NotValueOfType{reflect.Uint32, "5000000000"},
			&cptest.NotValueOfType{reflect.Uint64, "20000000000000000000"},
		}

		td.Cmp(t, errs.Errors, td.Bag(td.Flatten(wantErrs)))
		td.Cmp(t, target, want)
	})

	t.Run("float fields, no error", func(t *testing.T) {
		type structType struct {
			F32       float32
			Untouched int
			F64       float64
		}

		target := structType{1.0, 0, 2.0}

		kvm := map[string]string{
			"F32": "42.00",
			"F64": "1e9",
		}

		want := structType{42.0, 0, 1e9}

		err := cptest.KVMapUnmarshal(kvm, &target)

		td.CmpNoError(t, err)
		td.Cmp(t, target, want)
	})

	t.Run("float fields, values out of range or bogus", func(t *testing.T) {
		type structType struct {
			F32       float32
			Untouched int
			F64       float64
		}

		target := structType{1.0, 0, 2.0}

		kvm := map[string]string{
			"F32": "3.402824E+38",
			"F64": "-2.7976931348623157E+308",
		}

		want := structType{1.0, 0, 2.0}

		errs := cptest.KVMapUnmarshal(kvm, &target).(*multierror.Error)

		wantErrs := []error{
			&cptest.NotValueOfType{reflect.Float32, "3.402824E+38"},
			&cptest.NotValueOfType{reflect.Float64, "-2.7976931348623157E+308"},
		}

		td.Cmp(t, errs.Errors, wantErrs)
		td.Cmp(t, target, want)
	})

	t.Run("report missing fields", func(t *testing.T) {
		target := struct{}{}

		kvm := map[string]string{
			"Foo": "42",
			"Bar": "ハロー",
			"":    "えっ？",
		}

		errs := cptest.KVMapUnmarshal(kvm, &target).(*multierror.Error)

		td.CmpError(t, errs)

		wantErrs := []error{
			&cptest.MissingFieldError{"Foo"},
			&cptest.MissingFieldError{"Bar"},
			&cptest.MissingFieldError{""},
		}

		td.Cmp(t, errs.Errors, td.Bag(td.Flatten(wantErrs)))
	})
}
