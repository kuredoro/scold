package scold_test

import (
	"reflect"
	"strings"
	"testing"
	"time"
	"unicode"

	"github.com/hashicorp/go-multierror"
	"github.com/kuredoro/scold"
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

		err := scold.StringMapUnmarshal(sm, &target)

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

		err := scold.StringMapUnmarshal(sm, &target)

		td.CmpNoError(t, err)
		td.Cmp(t, target, want)
	})

	t.Run("unmarshal only works on structs or pointers to them", func(t *testing.T) {
		sm := map[string]string{}

		td.CmpPanic(t, func() { _ = scold.StringMapUnmarshal(sm, 42) }, scold.ErrNotAStructLike)

		i := 42
		td.CmpPanic(t, func() { _ = scold.StringMapUnmarshal(sm, &i) }, scold.ErrNotAStructLike)

		td.CmpPanic(t, func() { _ = scold.StringMapUnmarshal(sm, "foo") }, scold.ErrNotAStructLike)

		str := "foo"
		td.CmpPanic(t, func() { _ = scold.StringMapUnmarshal(sm, &str) }, scold.ErrNotAStructLike)

		td.CmpPanic(t, func() { _ = scold.StringMapUnmarshal(sm, []int{1, 2, 3}) }, scold.ErrNotAStructLike)

		td.CmpPanic(t, func() { _ = scold.StringMapUnmarshal(sm, [...]int{1, 2, 3}) }, scold.ErrNotAStructLike)

		td.CmpPanic(t, func() { _ = scold.StringMapUnmarshal(sm, sm) }, scold.ErrNotAStructLike)

		td.CmpPanic(t, func() { _ = scold.StringMapUnmarshal(sm, func() {}) }, scold.ErrNotAStructLike)

		td.CmpPanic(t, func() { _ = scold.StringMapUnmarshal(sm, make(chan int)) }, scold.ErrNotAStructLike)

		// ---

		err := scold.StringMapUnmarshal(sm, struct{}{})
		td.CmpNoError(t, err)

		test := struct{}{}
		err = scold.StringMapUnmarshal(sm, &test)
		td.CmpNoError(t, err)
	})

	t.Run("report missing fields", func(t *testing.T) {
		target := struct{}{}

		sm := map[string]string{
			"Foo":    "42",
			"Bar":    "ハロー",
			"AGAIN?": "435",
		}

		errs := scold.StringMapUnmarshal(sm, &target).(*multierror.Error)

		td.CmpError(t, errs)

		wantErrs := []error{
			&scold.FieldError{"Foo", scold.ErrUnknownField},
			&scold.FieldError{"Bar", scold.ErrUnknownField},
			&scold.FieldError{"AGAIN?", scold.ErrUnknownField},
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

		err := scold.StringMapUnmarshal(sm, &target)

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

		errs := scold.StringMapUnmarshal(sm, &target).(*multierror.Error)

		wantErrs := []error{
			&scold.FieldError{"I8", &scold.NotValueOfTypeError{reflect.Int8.String(), "-129", nil}},
			&scold.FieldError{"I16", &scold.NotValueOfTypeError{reflect.Int16.String(), "40000", nil}},
			&scold.FieldError{"I32", &scold.NotValueOfTypeError{reflect.Int32.String(), "-3000000000", nil}},
			&scold.FieldError{"I64", &scold.NotValueOfTypeError{reflect.Int64.String(), "10000000000000000000", nil}},
			&scold.FieldError{"Ui", &scold.NotValueOfTypeError{reflect.Uint.String(), "０", nil}},
			&scold.FieldError{"U8", &scold.NotValueOfTypeError{reflect.Uint8.String(), "300", nil}},
			&scold.FieldError{"U16", &scold.NotValueOfTypeError{reflect.Uint16.String(), "67000", nil}},
			&scold.FieldError{"U32", &scold.NotValueOfTypeError{reflect.Uint32.String(), "5000000000", nil}},
			&scold.FieldError{"U64", &scold.NotValueOfTypeError{reflect.Uint64.String(), "20000000000000000000", nil}},
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

		err := scold.StringMapUnmarshal(sm, &target)

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

		errs := scold.StringMapUnmarshal(sm, &target).(*multierror.Error)

		wantErrs := []error{
			&scold.FieldError{"F32", &scold.NotValueOfTypeError{reflect.Float32.String(), "3.402824E+38", nil}},
			&scold.FieldError{"F64", &scold.NotValueOfTypeError{reflect.Float64.String(), "-2.7976931348623157E+308", nil}},
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

			err := scold.StringMapUnmarshal(sm, &target)
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

			errs := scold.StringMapUnmarshal(sm, &target).(*multierror.Error)
			td.Cmp(t, errs.Errors, []error{
				&scold.FieldError{"B", &scold.NotValueOfTypeError{reflect.Bool.String(), step.value, nil}},
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

			errs := scold.StringMapUnmarshal(sm, &target).(*multierror.Error)
			td.Cmp(t, errs.Errors, []error{
				&scold.FieldError{"B", &scold.NotValueOfTypeError{reflect.Bool.String(), step.value, nil}},
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

		errs := scold.StringMapUnmarshal(sm, &target).(*multierror.Error)

		td.CmpError(t, errs)

		wantErrs := []error{
			&scold.FieldError{"Foo", scold.ErrUnknownField},
			&scold.FieldError{"Bar", scold.ErrUnknownField},
			&scold.FieldError{"", scold.ErrUnknownField},
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

		err := scold.StringMapUnmarshal(sm, &target)

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

		td.CmpPanic(t, func() { _ = scold.StringMapUnmarshal(sm, &target1) }, &scold.NotTextUnmarshalableTypeError{Field: "Info", Type: reflect.Struct, TypeName: "struct { Age int }"})

		type InfoType struct{ Age int }

		target2 := struct {
			Info InfoType
		}{}

		td.CmpPanic(t, func() { _ = scold.StringMapUnmarshal(sm, &target2) }, &scold.NotTextUnmarshalableTypeError{Field: "Info", Type: reflect.Struct, TypeName: "scold_test.InfoType"})

		target3 := struct {
			Info *InfoType
		}{}

		td.CmpPanic(t, func() { _ = scold.StringMapUnmarshal(sm, &target3) }, &scold.NotTextUnmarshalableTypeError{Field: "Info", Type: reflect.Ptr, TypeName: "*scold_test.InfoType"})

		type Numbers []int

		target4 := struct {
			Info Numbers
		}{}

		td.CmpPanic(t, func() { _ = scold.StringMapUnmarshal(sm, &target4) }, &scold.NotTextUnmarshalableTypeError{Field: "Info", Type: reflect.Slice, TypeName: "scold_test.Numbers"})
	})

	t.Run("pointer to deserializable type that was allocated", func(t *testing.T) {
		target := struct {
			Dur *scold.PositiveDuration
		}{&scold.PositiveDuration{time.Second}}

		sm := map[string]string{
			"Dur": "5s",
		}

		err := scold.StringMapUnmarshal(sm, &target)

		td.CmpNoError(t, err)
		td.Cmp(t, target, struct{ Dur *scold.PositiveDuration }{&scold.PositiveDuration{5 * time.Second}})
	})

	t.Run("pointer to deserializable type that was NOT allocated should allocate it", func(t *testing.T) {
		target := struct {
			Dur *scold.PositiveDuration
		}{}

		sm := map[string]string{
			"Dur": "5s",
		}

		err := scold.StringMapUnmarshal(sm, &target)

		td.CmpNoError(t, err)
		td.Cmp(t, target, struct{ Dur *scold.PositiveDuration }{&scold.PositiveDuration{5 * time.Second}})
	})

	t.Run("plain deserializable type should be filled anyway", func(t *testing.T) {
		target := struct {
			Dur scold.PositiveDuration
		}{}

		sm := map[string]string{
			"Dur": "5s",
		}

		err := scold.StringMapUnmarshal(sm, &target)

		td.CmpNoError(t, err)
		td.Cmp(t, target, struct{ Dur scold.PositiveDuration }{scold.PositiveDuration{5 * time.Second}})
	})

	t.Run("error if user-defined type failed to parse input (non-pointer version)", func(t *testing.T) {
		target := struct {
			Dur scold.PositiveDuration
		}{}

		sm := map[string]string{
			"Dur": "5sus",
		}

		errs := scold.StringMapUnmarshal(sm, &target).(*multierror.Error)

		td.Cmp(t, errs.Errors, []error{
			&scold.FieldError{"Dur", &scold.NotValueOfTypeError{"PositiveDuration", "5sus", nil}},
		})
		td.Cmp(t, target, struct{ Dur scold.PositiveDuration }{})
	})

	t.Run("error if user-defined type failed to parse input (empty pointer version)", func(t *testing.T) {
		target := struct {
			Dur *scold.PositiveDuration
		}{}

		sm := map[string]string{
			"Dur": "5sus",
		}

		errs := scold.StringMapUnmarshal(sm, &target).(*multierror.Error)

		td.Cmp(t, errs.Errors, []error{
			&scold.FieldError{"Dur", &scold.NotValueOfTypeError{"PositiveDuration", "5sus", nil}},
		})
		td.Cmp(t, target, struct{ Dur *scold.PositiveDuration }{})
	})

	t.Run("error if user-defined type failed to parse input (valid pointer version)", func(t *testing.T) {
		dur := &scold.PositiveDuration{time.Second}
		target := struct {
			Dur *scold.PositiveDuration
		}{dur}

		sm := map[string]string{
			"Dur": "5sus",
		}

		errs := scold.StringMapUnmarshal(sm, &target).(*multierror.Error)

		td.Cmp(t, errs.Errors, []error{
			&scold.FieldError{"Dur", &scold.NotValueOfTypeError{"PositiveDuration", "5sus", nil}},
		})
		td.Cmp(t, target, struct{ Dur *scold.PositiveDuration }{dur})
	})

	t.Run("optional transform functions", func(t *testing.T) {
		type TestStruct struct {
			One            int
			Aword          string
			PascalCase     *scold.PositiveDuration
			ShouldNotMatch uint
		}

		target := TestStruct{}

		sm := map[string]string{
			"one":            "1",
			"a_word":         "woah",
			"pascal_case":    "5s",
			"ShouldNotMatch": "12",
		}

		want := TestStruct{
			One:            1,
			Aword:          "woah",
			PascalCase:     &scold.PositiveDuration{5 * time.Second},
			ShouldNotMatch: 0,
		}

		callCounts := make([]int, 3)

		// Used to explicitly reject pascal case map keys
		isPascalCase := func(s string) bool {
			return strings.Count(s, "_") == 0 && unicode.IsUpper([]rune(s)[0])
		}

		capitalize := func(s string) string {
			callCounts[0]++

			if isPascalCase(s) {
				return ""
			}

			if s == "" {
				return s
			}

			runes := []rune(s)
			runes[0] = unicode.ToUpper(runes[0])
			return string(runes)
		}

		capitalizeWithNoUnderscore := func(s string) string {
			callCounts[1]++

			if isPascalCase(s) {
				return ""
			}

			if s == "" {
				return s
			}

			r := []rune(s)
			out := 0
			for in := 0; in < len(r); in++ {
				if r[in] == '_' {
					continue
				}

				r[out] = r[in]
				out++
			}

			r[0] = unicode.ToUpper(r[0])

			return string(r[:out])
		}

		toPascalCase := func(s string) string {
			callCounts[2]++

			if isPascalCase(s) {
				return ""
			}

			if s == "" {
				return s
			}

			r := []rune(s)
			out := 0
			for in, capitalize := 0, true; in < len(r); in++ {
				if r[in] == '_' {
					capitalize = true
					continue
				}

				if capitalize {
					r[out] = unicode.ToUpper(r[in])
					capitalize = false
				} else {
					r[out] = r[in]
				}

				out++
			}

			return string(r[:out])
		}

		err := scold.StringMapUnmarshal(sm, &target, capitalize, capitalizeWithNoUnderscore, toPascalCase).(*multierror.Error)

		td.Cmp(t, err.Errors, td.Bag(td.Flatten([]error{&scold.FieldError{"ShouldNotMatch", scold.ErrUnknownField}})))
		td.Cmp(t, target, want, "correctly deserialized")
		td.Cmp(t, callCounts, []int{4, 3, 2}, "transformers called sequentially until success")
	})
}

func TestPositiveDuration(t *testing.T) {
	t.Run("empty string", func(t *testing.T) {
		dur := &scold.PositiveDuration{}
		err := dur.UnmarshalText([]byte(""))

		td.CmpError(t, err)
	})

	t.Run("zero nanoseconds", func(t *testing.T) {
		dur := &scold.PositiveDuration{}
		err := dur.UnmarshalText([]byte("0ns"))

		td.CmpNoError(t, err)
		td.Cmp(t, dur.Duration, 0*time.Nanosecond)
	})

	t.Run("one second", func(t *testing.T) {
		dur := &scold.PositiveDuration{}
		err := dur.UnmarshalText([]byte("1s"))

		td.CmpNoError(t, err)
		td.Cmp(t, dur.Duration, time.Second)
	})

	t.Run("one second without suffix", func(t *testing.T) {
		dur := &scold.PositiveDuration{}
		err := dur.UnmarshalText([]byte("1"))

		td.Cmp(t, err, scold.ErrDurationWithoutSuffix)
		td.Cmp(t, dur.Duration, time.Second)
	})

	t.Run("ten seconds", func(t *testing.T) {
		dur := &scold.PositiveDuration{}
		err := dur.UnmarshalText([]byte("10s"))

		td.CmpNoError(t, err)
		td.Cmp(t, dur.Duration, 10*time.Second)
	})

	t.Run("ten and half seconds without suffix", func(t *testing.T) {
		dur := &scold.PositiveDuration{}
		err := dur.UnmarshalText([]byte("10.5"))

		td.Cmp(t, err, scold.ErrDurationWithoutSuffix)
		td.Cmp(t, dur.Duration, 10500*time.Millisecond)
	})

	t.Run("1 and half milliseconds", func(t *testing.T) {
		dur := &scold.PositiveDuration{}
		err := dur.UnmarshalText([]byte("1.5ms"))

		td.CmpNoError(t, err)
		td.Cmp(t, dur.Duration, 1500*time.Microsecond)
	})

	t.Run("negative second is forbidden", func(t *testing.T) {
		dur := &scold.PositiveDuration{}
		err := dur.UnmarshalText([]byte("-1s"))

		td.Cmp(t, err, scold.ErrNegativePositiveDuration)
	})

	t.Run("minus ten and half seconds without suffix", func(t *testing.T) {
		dur := &scold.PositiveDuration{}
		err := dur.UnmarshalText([]byte("-10.5"))

		td.Cmp(t, err, scold.ErrNegativePositiveDuration)
	})

	t.Run("negative fractional duration is forbidden", func(t *testing.T) {
		dur := &scold.PositiveDuration{}
		err := dur.UnmarshalText([]byte("-42.0us"))

		td.Cmp(t, err, scold.ErrNegativePositiveDuration)
	})

	t.Run("jibberish is forbidden", func(t *testing.T) {
		dur := &scold.PositiveDuration{}
		err := dur.UnmarshalText([]byte("ko-hi-"))

		td.Cmp(t, err, scold.ErrDurationBadSyntax)
	})
}
