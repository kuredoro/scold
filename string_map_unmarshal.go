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
	ErrNotAStructLike = StringError("not a struct-like")

	ErrUnknownField = StringError("unknown field")
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

// FromStringer should not alter the its value, if error occurs.
type FromStringer interface {
	FromString(string) error
}

type FieldError struct {
	FieldName string
	Err       error
}

func (e *FieldError) Error() string {
	return fmt.Sprintf("field %q: %v", e.FieldName, e.Err)
}

func (e *FieldError) Unwrap() error {
	return e.Err
}

type NotValueOfTypeError struct {
	Type  string
	Value string
}

func (e *NotValueOfTypeError) Error() string {
	return fmt.Sprintf("value %q doesn't match %v type", e.Value, e.Type)
}

func (e *NotValueOfTypeError) Equal(other *NotValueOfTypeError) bool {
	return e.Type == other.Type && e.Value == other.Value
}

type NotStringUnmarshalableTypeError struct {
	Field    string
	Type     reflect.Kind
	TypeName string
}

func (e *NotStringUnmarshalableTypeError) Error() string {
	return fmt.Sprintf("field %q is of type %v (%v) and cannot be unmarshaled from string, because it is not of fundamental type or because the type doesn't implement FromString(string) method", e.Field, e.TypeName, e.Type)
}

func (e *NotStringUnmarshalableTypeError) Equal(other *NotStringUnmarshalableTypeError) bool {
	return e.Field == other.Field && e.Type == other.Type && e.TypeName == other.TypeName
}

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

		// A user-defined type should satisfy FromStringer
		var unmarshalDest FromStringer
		var isDeserializable bool
		if field.Kind() == reflect.Ptr {
			unmarshalDest, isDeserializable = field.Interface().(FromStringer)
		} else {
			unmarshalDest, isDeserializable = field.Addr().Interface().(FromStringer)
		}

		if isDeserializable {
			// Make sure that the field is allocated, so that the FromString
			// method had a place to write back to. But also remember if we
			// changed the field, so to revert it back if error occurs.
			wasNil := false
			if field.Kind() == reflect.Ptr && field.IsNil() {
				field.Set(reflect.New(field.Type().Elem()))
				unmarshalDest = field.Interface().(FromStringer)
				wasNil = true
			}

			err := unmarshalDest.FromString(v)

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
			panic(&NotStringUnmarshalableTypeError{Field: k, Type: field.Kind(), TypeName: fmt.Sprintf("%T", field.Interface())})
		}
	}

	return errs.ErrorOrNil()
}
