package cptest

import (
	"fmt"
	"reflect"
    "strconv"

	"github.com/hashicorp/go-multierror"
)

const(
    // NotStructLike is issued when a destination type doesn't behave like
    // a struct. struct and *struct types have the same syntax for manipulating
    // them, so they are considered struct-like.
    NotAStructLike = StringError("not a struct-like")
)

type MissingFieldError struct {
    FieldName string
}

func (e *MissingFieldError) Error() string {
    return fmt.Sprintf("field %q doesn't exist", e.FieldName)
}

type NotValueOfType struct {
    Type reflect.Kind
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

        if field.Kind() == reflect.Int {
            parsed, err := strconv.Atoi(v)
            if err != nil {
                errs = multierror.Append(errs, &NotValueOfType{field.Kind(), v})
                continue
            }
            field.Set(reflect.ValueOf(parsed))
        } else if field.Kind() == reflect.Int8 {
            parsed, err := strconv.ParseInt(v, 10, 8)
            if err != nil {
                errs = multierror.Append(errs, &NotValueOfType{field.Kind(), v})
                continue
            }
            field.Set(reflect.ValueOf(int8(parsed)))
        } else if field.Kind() == reflect.Int16 {
            parsed, err := strconv.ParseInt(v, 10, 16)
            if err != nil {
                errs = multierror.Append(errs, &NotValueOfType{field.Kind(), v})
                continue
            }
            field.Set(reflect.ValueOf(int16(parsed)))
        } else if field.Kind() == reflect.Int32 {
            parsed, err := strconv.ParseInt(v, 10, 32)
            if err != nil {
                errs = multierror.Append(errs, &NotValueOfType{field.Kind(), v})
                continue
            }
            field.Set(reflect.ValueOf(int32(parsed)))
        } else if field.Kind() == reflect.Int64 {
            parsed, err := strconv.ParseInt(v, 10, 64)
            if err != nil {
                errs = multierror.Append(errs, &NotValueOfType{field.Kind(), v})
                continue
            }
            field.Set(reflect.ValueOf(int64(parsed)))
        } else if field.Kind() == reflect.Uint {
            parsed, err := strconv.ParseUint(v, 10, 0)
            if err != nil {
                errs = multierror.Append(errs, &NotValueOfType{field.Kind(), v})
                continue
            }
            field.Set(reflect.ValueOf(uint(parsed)))
        } else if field.Kind() == reflect.Uint8 {
            parsed, err := strconv.ParseUint(v, 10, 8)
            if err != nil {
                errs = multierror.Append(errs, &NotValueOfType{field.Kind(), v})
                continue
            }
            field.Set(reflect.ValueOf(uint8(parsed)))
        } else if field.Kind() == reflect.Uint16 {
            parsed, err := strconv.ParseUint(v, 10, 16)
            if err != nil {
                errs = multierror.Append(errs, &NotValueOfType{field.Kind(), v})
                continue
            }
            field.Set(reflect.ValueOf(uint16(parsed)))
        } else if field.Kind() == reflect.Uint32 {
            parsed, err := strconv.ParseUint(v, 10, 32)
            if err != nil {
                errs = multierror.Append(errs, &NotValueOfType{field.Kind(), v})
                continue
            }
            field.Set(reflect.ValueOf(uint32(parsed)))
        } else if field.Kind() == reflect.Uint64 {
            parsed, err := strconv.ParseUint(v, 10, 64)
            if err != nil {
                errs = multierror.Append(errs, &NotValueOfType{field.Kind(), v})
                continue
            }
            field.Set(reflect.ValueOf(uint64(parsed)))
        } else {
            // ...
        }
    }

    return errs.ErrorOrNil()
}
