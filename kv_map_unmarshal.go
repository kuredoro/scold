package cptest

import (
	"fmt"
	"reflect"
	"strconv"

	"github.com/hashicorp/go-multierror"
)

const (
	// NotStructLike is issued when a destination type doesn't behave like
	// a struct. struct and *struct types have the same syntax for manipulating
	// them, so they are considered struct-like.
	NotAStructLike = StringError("not a struct-like")
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

type MissingFieldError struct {
	FieldName string
}

func (e *MissingFieldError) Error() string {
	return fmt.Sprintf("field %q doesn't exist", e.FieldName)
}

type NotValueOfType struct {
	Type   string
	Value  string
	Reason error
}

func (e *NotValueOfType) Error() string {
	return fmt.Sprintf("value %q doesn't match %v type", e.Value, e.Type)
}

func (e *NotValueOfType) Equal(other *NotValueOfType) bool {
	return e.Type == other.Type && e.Value == other.Value
}

func (e *NotValueOfType) Unwrap() error {
    return e.Reason
}

type NotStringUnmarshalableType struct {
	Field    string
	Type     reflect.Kind
	TypeName string
}

func (e *NotStringUnmarshalableType) Error() string {
	return fmt.Sprintf("field %q is of type %v (%v) and cannot be unmarshaled from string, because it is not of fundamental type or because the type doesn't implement FromString(string) method", e.Field, e.TypeName, e.Type)
}

func (e *NotStringUnmarshalableType) Equal(other *NotStringUnmarshalableType) bool {
	return e.Field == other.Field && e.Type == other.Type && e.TypeName == other.TypeName
}

type KVMap = map[string]string

func KVMapUnmarshal(kvm KVMap, data interface{}) error {
	val := reflect.ValueOf(data)

	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	if val.Kind() != reflect.Struct {
		panic(NotAStructLike)
	}

	var errs *multierror.Error

	for k, v := range kvm {
		field := val.FieldByName(k)

		if !field.IsValid() {
			errs = multierror.Append(errs, &MissingFieldError{k})
			continue
		}

		// A user-defined type should contain FromString method
        // XXX: Should do via interfaces
        fromStringMethod := field.MethodByName("FromString")
        if !fromStringMethod.IsValid() {
            fromStringMethod = field.Addr().MethodByName("FromString")

            // If we found conversion method on a pointer, then we should work
            // on the pointer to the field further.
            if fromStringMethod.IsValid() {
                field = field.Addr()
            }
        }

        if fromStringMethod.IsValid() {
            // Make sure that the field is allocated, so that the FromString
            // method had a place to write back to. But also remember if we
            // changed the field, so to revert it back if error occurs.
            wasNil := false
            if field.IsNil() {
                field.Set(reflect.New(field.Type().Elem()))
                wasNil = true
            }

			output := fromStringMethod.Call([]reflect.Value{reflect.ValueOf(v)})

			if !output[0].IsNil() {
				err := output[0].Interface().(error)
				errs = multierror.Append(errs, &NotValueOfType{field.Type().Elem().Name(), v, err})

                if wasNil {
                    field.Set(reflect.Zero(field.Type()))
                }
			}
			continue
		}

		if bitSize, found := intParsers[field.Kind()]; found {
			parsed, err := strconv.ParseInt(v, 10, bitSize)
			if err != nil {
				errs = multierror.Append(errs, &NotValueOfType{field.Kind().String(), v, err})
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
				errs = multierror.Append(errs, &NotValueOfType{field.Kind().String(), v, err})
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
				errs = multierror.Append(errs, &NotValueOfType{field.Kind().String(), v, err})
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
				errs = multierror.Append(errs, &NotValueOfType{field.Kind().String(), v, err})
				continue
			}

			field.Set(reflect.ValueOf(parsed))
		} else if field.Kind() == reflect.String {
			field.Set(reflect.ValueOf(v))
		} else {
			panic(&NotStringUnmarshalableType{Field: k, Type: field.Kind(), TypeName: field.Type().Name()})
		}
	}

	return errs.ErrorOrNil()
}
