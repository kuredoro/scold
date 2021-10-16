package cptest

import (
	"encoding"
	"fmt"
	"reflect"
	"strconv"
	"time"

	"github.com/hashicorp/go-multierror"
)

const (
	// ErrNotAStructLike is issued when a destination type doesn't behave like
	// a struct. struct and *struct types have the same syntax for manipulating
	// them, so they are considered struct-like.
	ErrNotAStructLike = StringError("not a struct-like")

	// ErrUnknownField is issued when the deserialization destination cannot be
	// found in a struct-like via reflection.
	ErrUnknownField = StringError("unknown field")

	// ErrNegativePositiveDuration is issued when PositiveDuration is attempted
	// to be initialized with negative value.
	ErrNegativePositiveDuration = StringError("PositiveDuration accepts only positive durations")
)

var intParsers = map[reflect.Kind]int{
	reflect.Int:   0,
	reflect.Int8:  8,
	reflect.Int16: 16,
	reflect.Int32: 32,
	reflect.Int64: 64,
}

var uintParsers = map[reflect.Kind]int{
	reflect.Uint:   0,
	reflect.Uint8:  8,
	reflect.Uint16: 16,
	reflect.Uint32: 32,
	reflect.Uint64: 64,
}

var floatParsers = map[reflect.Kind]int{
	reflect.Float32: 32,
	reflect.Float64: 64,
}

var complexParsers = map[reflect.Kind]int{
	reflect.Complex64:  64,
	reflect.Complex128: 128,
}

// PositiveDuration is a wrapper around time.Duration that allows
// StringMapsUnmarshal to parse it from a string and that forbids negative
// durations. Implements encoding.TextUnmarshaler.
type PositiveDuration struct{ time.Duration }

// NewPositiveDuration returns a PositiveDuration with the specified value.
// Panics if value is negative.
func NewPositiveDuration(dur time.Duration) PositiveDuration {
	if dur < 0 {
		panic(ErrNegativePositiveDuration)
	}

	return PositiveDuration{dur}
}

// UnmarshalText will delegate parsing to built-in time.ParseDuration and,
// hence, accept the same format as time.ParseDuration. It will also
// reject negative durations.
func (d *PositiveDuration) UnmarshalText(b []byte) error {
	dur, err := time.ParseDuration(string(b))
	if dur.Nanoseconds() < 0 {
		return ErrNegativePositiveDuration
	}

    if err != nil {
        num, err := strconv.ParseFloat(string(b), 64)
        if err == nil {
            *d = PositiveDuration{time.Duration(num * float64(time.Second))}
            return nil
        }
    }

	*d = PositiveDuration{dur}
	return err
}

// StringMapUnmarshal accepts a string map and for each key-value pair tries
// to find an identically named field in the provided object, parse the
// string value according to the field's type and assign the parsed value
// to it.
//
// The field's type should be: int (any flavor), uint (any flavor), string,
// bool, any struct or a pointer to a struct. The structs should implement
// encoding.TextUnmarshaler standard interface to be parsed by this function,
// otherwise NotUnmarshalableTypeError is issued.
//
// If the destination object is not struct or a pointer to a struct,
// ErrNotAStructLike is issued.
//
// If a map's key cannot be mapped to a field within the destination object,
// FieldError wrapping ErrUnknownField is issued.
//
// If a map's value cannot be parsed to the destination type, a FieldError
// wrapping a NotValueOfTypeError is issued.
//
// StringMapUnmarshal makes sure that if any error occurs during unmarshaling
// of a field, the field's previous value is retained.
//
// This function will accumulate all produced errors and return an instance of
// *multierror.Error type.
//
// Since a key must match the field's name perfectly, and this function operates
// only on exported fields, this would mean that only capitalized keys would
// be accepted. This may not be desired. An array of transformers can be
// supplied to change the field matching behavior. Each map's key will be
// fed to one transformer at a time in the order they were passed to this
// function, until a match is found. The transformer's job, then, is to
// convert an arbitrary string to a possible exported field's name in the
// destination object. If a transformer succeeds, the successive transformers
// are not applied. If the field still could not be found, a FieldError
// wrapping ErrUnknownField is issued.
func StringMapUnmarshal(kvm map[string]string, data interface{}, transformers ...func(string) string) error {
	val := reflect.ValueOf(data)

	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	if val.Kind() != reflect.Struct {
		panic(ErrNotAStructLike)
	}

	var errs *multierror.Error

	for k, v := range kvm {
		var field reflect.Value
		if len(transformers) == 0 {
			field = val.FieldByName(k)
		} else {
			for _, transformer := range transformers {
				nameCandidate := transformer(k)
				field = val.FieldByName(nameCandidate)

				if field.IsValid() {
					break
				}
			}
		}

		if !field.IsValid() {
			errs = multierror.Append(errs, &FieldError{k, ErrUnknownField})
			continue
		}

		// A user-defined type should satisfy encoding.TextUnmarshaler
		var unmarshalDest encoding.TextUnmarshaler
		var isDeserializable bool
		if field.Kind() == reflect.Ptr {
			unmarshalDest, isDeserializable = field.Interface().(encoding.TextUnmarshaler)
		} else {
			unmarshalDest, isDeserializable = field.Addr().Interface().(encoding.TextUnmarshaler)
		}

		if isDeserializable {
			// Make sure that the field is allocated, so that the FromString
			// method had a place to write back to. But also remember if we
			// changed the field, so to revert it back if error occurs.
			wasNil := false
			if field.Kind() == reflect.Ptr && field.IsNil() {
				field.Set(reflect.New(field.Type().Elem()))
				unmarshalDest = field.Interface().(encoding.TextUnmarshaler)
				wasNil = true
			}

			err := unmarshalDest.UnmarshalText([]byte(v))

			if err != nil {
				var typeName string
				if field.Kind() == reflect.Ptr {
					typeName = field.Type().Elem().Name()
				} else {
					typeName = field.Type().Name()
				}

				errs = multierror.Append(errs, &FieldError{k, &NotValueOfTypeError{typeName, v}})

				if wasNil {
					field.Set(reflect.Zero(field.Type()))
				}
			}
			continue
		}

		if bitSize, found := intParsers[field.Kind()]; found {
			parsed, err := strconv.ParseInt(v, 10, bitSize)
			if err != nil {
				errs = multierror.Append(errs, &FieldError{k, &NotValueOfTypeError{field.Kind().String(), v}})
				continue
			}

			// This is the minimum I succeeded to decrease the boilerplate...
			if field.Kind() == reflect.Int {
				field.Set(reflect.ValueOf(int(parsed)))
			} else if field.Kind() == reflect.Int8 {
				field.Set(reflect.ValueOf(int8(parsed)))
			} else if field.Kind() == reflect.Int16 {
				field.Set(reflect.ValueOf(int16(parsed)))
			} else if field.Kind() == reflect.Int32 {
				field.Set(reflect.ValueOf(int32(parsed)))
			} else if field.Kind() == reflect.Int64 {
				field.Set(reflect.ValueOf(int64(parsed)))
			}
		} else if bitSize, found := uintParsers[field.Kind()]; found {
			parsed, err := strconv.ParseUint(v, 10, bitSize)
			if err != nil {
				errs = multierror.Append(errs, &FieldError{k, &NotValueOfTypeError{field.Kind().String(), v}})
				continue
			}

			if field.Kind() == reflect.Uint {
				field.Set(reflect.ValueOf(uint(parsed)))
			} else if field.Kind() == reflect.Uint8 {
				field.Set(reflect.ValueOf(uint8(parsed)))
			} else if field.Kind() == reflect.Uint16 {
				field.Set(reflect.ValueOf(uint16(parsed)))
			} else if field.Kind() == reflect.Uint32 {
				field.Set(reflect.ValueOf(uint32(parsed)))
			} else if field.Kind() == reflect.Uint64 {
				field.Set(reflect.ValueOf(uint64(parsed)))
			}
		} else if bitSize, found := floatParsers[field.Kind()]; found {
			parsed, err := strconv.ParseFloat(v, bitSize)
			if err != nil {
				errs = multierror.Append(errs, &FieldError{k, &NotValueOfTypeError{field.Kind().String(), v}})
				continue
			}

			if field.Kind() == reflect.Float32 {
				field.Set(reflect.ValueOf(float32(parsed)))
			} else if field.Kind() == reflect.Float64 {
				field.Set(reflect.ValueOf(float64(parsed)))
			}
		} else if _, found := complexParsers[field.Kind()]; found {
			panic("Unmarshaling complex numbers from strings is not implemented")
		} else if field.Kind() == reflect.Bool {
			parsed, err := strconv.ParseBool(v)
			if err != nil {
				errs = multierror.Append(errs, &FieldError{k, &NotValueOfTypeError{field.Kind().String(), v}})
				continue
			}

			field.Set(reflect.ValueOf(parsed))
		} else if field.Kind() == reflect.String {
			field.Set(reflect.ValueOf(v))
		} else {
			panic(&NotTextUnmarshalableTypeError{Field: k, Type: field.Kind(), TypeName: fmt.Sprintf("%T", field.Interface())})
		}
	}

	return errs.ErrorOrNil()
}
