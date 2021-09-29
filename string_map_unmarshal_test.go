package cptest_test

import (
	"reflect"
	"testing"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/kuredoro/cptest"
	"github.com/maxatome/go-testdeep/td"
)

func init() {
	td.DefaultContextConfig.UseEqual = true
	td.DefaultContextConfig.MaxErrors = -1
}

func TestStringAttributesUnmarshal(t *testing.T) {
	t.Run("empty map and struct", func(t *testing.T) {
		var target struct{}

		sm := map[string]string{} // String Map

		err := cptest.StringMapUnmarshal(sm, &target)

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

		sm := map[string]string{}

		want := target

		err := cptest.StringMapUnmarshal(sm, &target)

		td.CmpNoError(t, err)
		td.Cmp(t, target, want)
	})

	t.Run("unmarshal only works on structs or pointers to them", func(t *testing.T) {
		sm := map[string]string{}

		td.CmpPanic(t, func() { _ = cptest.StringMapUnmarshal(sm, 42) }, cptest.ErrNotAStructLike)

		i := 42
		td.CmpPanic(t, func() { _ = cptest.StringMapUnmarshal(sm, &i) }, cptest.ErrNotAStructLike)

		td.CmpPanic(t, func() { _ = cptest.StringMapUnmarshal(sm, "foo") }, cptest.ErrNotAStructLike)

		str := "foo"
		td.CmpPanic(t, func() { _ = cptest.StringMapUnmarshal(sm, &str) }, cptest.ErrNotAStructLike)

		td.CmpPanic(t, func() { _ = cptest.StringMapUnmarshal(sm, []int{1, 2, 3}) }, cptest.ErrNotAStructLike)

		td.CmpPanic(t, func() { _ = cptest.StringMapUnmarshal(sm, [...]int{1, 2, 3}) }, cptest.ErrNotAStructLike)

		td.CmpPanic(t, func() { _ = cptest.StringMapUnmarshal(sm, sm) }, cptest.ErrNotAStructLike)

		td.CmpPanic(t, func() { _ = cptest.StringMapUnmarshal(sm, func() {}) }, cptest.ErrNotAStructLike)

		td.CmpPanic(t, func() { _ = cptest.StringMapUnmarshal(sm, make(chan int)) }, cptest.ErrNotAStructLike)

		// ---

		err := cptest.StringMapUnmarshal(sm, struct{}{})
		td.CmpNoError(t, err)

		test := struct{}{}
		err = cptest.StringMapUnmarshal(sm, &test)
		td.CmpNoError(t, err)
	})

	t.Run("report missing fields", func(t *testing.T) {
		target := struct{}{}

		sm := map[string]string{
			"Foo":    "42",
			"Bar":    "ハロー",
			"AGAIN?": "435",
		}

		errs := cptest.StringMapUnmarshal(sm, &target).(*multierror.Error)

		td.CmpError(t, errs)

		wantErrs := []error{
			&cptest.FieldError{"Foo", cptest.ErrUnknownField},
			&cptest.FieldError{"Bar", cptest.ErrUnknownField},
			&cptest.FieldError{"AGAIN?", cptest.ErrUnknownField},
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

		sm := map[string]string{
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

		err := cptest.StringMapUnmarshal(sm, &target)

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

		sm := map[string]string{
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

		errs := cptest.StringMapUnmarshal(sm, &target).(*multierror.Error)

		wantErrs := []error{
			&cptest.FieldError{"I8", &cptest.NotValueOfTypeError{reflect.Int8.String(), "-129"}},
			&cptest.FieldError{"I16", &cptest.NotValueOfTypeError{reflect.Int16.String(), "40000"}},
			&cptest.FieldError{"I32", &cptest.NotValueOfTypeError{reflect.Int32.String(), "-3000000000"}},
			&cptest.FieldError{"I64", &cptest.NotValueOfTypeError{reflect.Int64.String(), "10000000000000000000"}},
			&cptest.FieldError{"Ui", &cptest.NotValueOfTypeError{reflect.Uint.String(), "０"}},
			&cptest.FieldError{"U8", &cptest.NotValueOfTypeError{reflect.Uint8.String(), "300"}},
			&cptest.FieldError{"U16", &cptest.NotValueOfTypeError{reflect.Uint16.String(), "67000"}},
			&cptest.FieldError{"U32", &cptest.NotValueOfTypeError{reflect.Uint32.String(), "5000000000"}},
			&cptest.FieldError{"U64", &cptest.NotValueOfTypeError{reflect.Uint64.String(), "20000000000000000000"}},
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

		sm := map[string]string{
			"F32": "42.00",
			"F64": "1e9",
		}

		want := structType{42.0, 0, 1e9}

		err := cptest.StringMapUnmarshal(sm, &target)

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

		sm := map[string]string{
			"F32": "3.402824E+38",
			"F64": "-2.7976931348623157E+308",
		}

		want := structType{1.0, 0, 2.0}

		errs := cptest.StringMapUnmarshal(sm, &target).(*multierror.Error)

		wantErrs := []error{
			&cptest.FieldError{"F32", &cptest.NotValueOfTypeError{reflect.Float32.String(), "3.402824E+38"}},
			&cptest.FieldError{"F64", &cptest.NotValueOfTypeError{reflect.Float64.String(), "-2.7976931348623157E+308"}},
		}

		td.Cmp(t, errs.Errors, td.Bag(td.Flatten(wantErrs)))
		td.Cmp(t, target, want)
	})

	t.Run("bool field, no error", func(t *testing.T) {
		type structType struct {
			B bool
		}

		target := structType{false}

		sequence := []struct {
			value string
			want  structType
		}{
			{"true", structType{true}},
			{"false", structType{false}},
			{"1", structType{true}},
			{"0", structType{false}},
			{"T", structType{true}},
			{"F", structType{false}},
			{"t", structType{true}},
			{"f", structType{false}},
			{"True", structType{true}},
			{"False", structType{false}},
			{"TRUE", structType{true}},
			{"FALSE", structType{false}},
			{"t", structType{true}},
			{"f", structType{false}},
		}

		for _, step := range sequence {
			sm := map[string]string{
				"B": step.value,
			}

			err := cptest.StringMapUnmarshal(sm, &target)
			td.CmpNoError(t, err)
			td.Cmp(t, target, step.want, "for value %q", step.value)
		}
	})

	t.Run("bool field, bogus values", func(t *testing.T) {
		type structType struct {
			B bool
		}

		target := structType{false}

		seq1 := []struct {
			value string
			want  structType
		}{
			{"tRuE", structType{false}},
			{"2", structType{false}},
			{"~OwO~", structType{false}},
		}

		for _, step := range seq1 {
			sm := map[string]string{
				"B": step.value,
			}

			errs := cptest.StringMapUnmarshal(sm, &target).(*multierror.Error)
			td.Cmp(t, errs.Errors, []error{
				&cptest.FieldError{"B", &cptest.NotValueOfTypeError{reflect.Bool.String(), step.value}},
			})
			td.Cmp(t, target, step.want, "for value %q", step.value)
		}

		target = structType{true}

		seq2 := []struct {
			value string
			want  structType
		}{
			{"fAlSe", structType{true}},
			{"00", structType{true}},
			{"x_x", structType{true}},
		}

		for _, step := range seq2 {
			sm := map[string]string{
				"B": step.value,
			}

			errs := cptest.StringMapUnmarshal(sm, &target).(*multierror.Error)
			td.Cmp(t, errs.Errors, []error{
				&cptest.FieldError{"B", &cptest.NotValueOfTypeError{reflect.Bool.String(), step.value}},
			})
			td.Cmp(t, target, step.want, "for value %q", step.value)
		}
	})

	t.Run("report missing fields", func(t *testing.T) {
		target := struct{}{}

		sm := map[string]string{
			"Foo": "42",
			"Bar": "ハロー",
			"":    "えっ？",
		}

		errs := cptest.StringMapUnmarshal(sm, &target).(*multierror.Error)

		td.CmpError(t, errs)

		wantErrs := []error{
			&cptest.FieldError{"Foo", cptest.ErrUnknownField},
			&cptest.FieldError{"Bar", cptest.ErrUnknownField},
			&cptest.FieldError{"", cptest.ErrUnknownField},
		}

		td.Cmp(t, errs.Errors, td.Bag(td.Flatten(wantErrs)))
	})

	t.Run("string fields", func(t *testing.T) {
		type testType struct {
			Foo       string
			Bar       string
			Zap       string
			Untouched string
		}

		target := testType{}

		sm := map[string]string{
			"Foo": "42",
			"Bar": "ハロー",
			"Zap": "",
		}

		err := cptest.StringMapUnmarshal(sm, &target)

		td.CmpNoError(t, err)

		td.Cmp(t, target, testType{"42", "ハロー", "", ""})
	})

	t.Run("plain struct fields cause panic", func(t *testing.T) {
		target1 := struct {
			Info struct{ Age int }
		}{}

		sm := map[string]string{
			"Info": "my age is over 9000",
		}

		td.CmpPanic(t, func() { _ = cptest.StringMapUnmarshal(sm, &target1) }, &cptest.NotStringUnmarshalableTypeError{Field: "Info", Type: reflect.Struct, TypeName: "struct { Age int }"})

		type InfoType struct{ Age int }

		target2 := struct {
			Info InfoType
		}{}

		td.CmpPanic(t, func() { _ = cptest.StringMapUnmarshal(sm, &target2) }, &cptest.NotStringUnmarshalableTypeError{Field: "Info", Type: reflect.Struct, TypeName: "cptest_test.InfoType"})

		target3 := struct {
			Info *InfoType
		}{}

		td.CmpPanic(t, func() { _ = cptest.StringMapUnmarshal(sm, &target3) }, &cptest.NotStringUnmarshalableTypeError{Field: "Info", Type: reflect.Ptr, TypeName: "*cptest_test.InfoType"})

		type Numbers []int

		target4 := struct {
			Info Numbers
		}{}

		td.CmpPanic(t, func() { _ = cptest.StringMapUnmarshal(sm, &target4) }, &cptest.NotStringUnmarshalableTypeError{Field: "Info", Type: reflect.Slice, TypeName: "cptest_test.Numbers"})
	})

	t.Run("pointer to deserializable type that was allocated", func(t *testing.T) {
		target := struct {
			Dur *duration
		}{&duration{time.Second}}

		sm := map[string]string{
			"Dur": "5s",
		}

		err := cptest.StringMapUnmarshal(sm, &target)

		td.CmpNoError(t, err)
		td.Cmp(t, target, struct{ Dur *duration }{&duration{5 * time.Second}})
	})

	t.Run("pointer to deserializable type that was NOT allocated should allocate it", func(t *testing.T) {
		target := struct {
			Dur *duration
		}{}

		sm := map[string]string{
			"Dur": "5s",
		}

		err := cptest.StringMapUnmarshal(sm, &target)

		td.CmpNoError(t, err)
		td.Cmp(t, target, struct{ Dur *duration }{&duration{5 * time.Second}})
	})

	t.Run("plain deserializable type should be filled anyway", func(t *testing.T) {
		target := struct {
			Dur duration
		}{}

		sm := map[string]string{
			"Dur": "5s",
		}

		err := cptest.StringMapUnmarshal(sm, &target)

		td.CmpNoError(t, err)
		td.Cmp(t, target, struct{ Dur duration }{duration{5 * time.Second}})
	})

	t.Run("error if user-defined type failed to parse input (non-pointer version)", func(t *testing.T) {
		target := struct {
			Dur duration
		}{}

		sm := map[string]string{
			"Dur": "5sus",
		}

		errs := cptest.StringMapUnmarshal(sm, &target).(*multierror.Error)

		td.Cmp(t, errs.Errors, []error{
			&cptest.FieldError{"Dur", &cptest.NotValueOfTypeError{"duration", "5sus"}},
		})
		td.Cmp(t, target, struct{ Dur duration }{})
	})

	t.Run("error if user-defined type failed to parse input (empty pointer version)", func(t *testing.T) {
		target := struct {
			Dur *duration
		}{}

		sm := map[string]string{
			"Dur": "5sus",
		}

		errs := cptest.StringMapUnmarshal(sm, &target).(*multierror.Error)

		td.Cmp(t, errs.Errors, []error{
			&cptest.FieldError{"Dur", &cptest.NotValueOfTypeError{"duration", "5sus"}},
		})
		td.Cmp(t, target, struct{ Dur *duration }{})
	})

	t.Run("error if user-defined type failed to parse input (valid pointer version)", func(t *testing.T) {
		dur := &duration{time.Second}
		target := struct {
			Dur *duration
		}{dur}

		sm := map[string]string{
			"Dur": "5sus",
		}

		errs := cptest.StringMapUnmarshal(sm, &target).(*multierror.Error)

		td.Cmp(t, errs.Errors, []error{
			&cptest.FieldError{"Dur", &cptest.NotValueOfTypeError{"duration", "5sus"}},
		})
		td.Cmp(t, target, struct{ Dur *duration }{dur})
	})
}

type duration struct{ time.Duration }

func (d *duration) FromString(str string) error {
	dur, err := time.ParseDuration(str)
	*d = duration{dur}
	return err
}
