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

type MissingFieldError struct {
	FieldName string
}

func (e *MissingFieldError) Error() string {
	return fmt.Sprintf("field %q doesn't exist", e.FieldName)
}

type NotValueOfType struct {
	Type  reflect.Kind
	Value string
}

func (e *NotValueOfType) Error() string {
	return fmt.Sprintf("value %q doesn't match %v type", e.Value, e.Type)
}

func (e *NotValueOfType) Equal(other *NotValueOfType) bool {
	return e.Type == other.Type && e.Value == other.Value
}

type KVMap = map[string]string

func KVMapUnmarshal(kvm KVMap, data interface{}) error {
	val := reflect.ValueOf(data)

	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	if val.Kind() != reflect.Struct {
		return NotAStructLike
	}

	var errs *multierror.Error

	for k, v := range kvm {
		// What if k empty?
		field := val.FieldByName(k)

		if !field.IsValid() {
			errs = multierror.Append(errs, &MissingFieldError{k})
			continue
		}

		if bitSize, found := intParsers[field.Kind()]; found {
			parsed, err := strconv.ParseInt(v, 10, bitSize)
			if err != nil {
				errs = multierror.Append(errs, &NotValueOfType{field.Kind(), v})
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
				errs = multierror.Append(errs, &NotValueOfType{field.Kind(), v})
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
		} else {
			// ...
		}
	}

	return errs.ErrorOrNil()
}
