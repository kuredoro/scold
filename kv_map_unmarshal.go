package cptest

import (
	"fmt"
	"reflect"

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
    return fmt.Sprintf("struct field %q doesn't exist", e.FieldName)
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

    for k := range kvm {
        // What if k empty?
        field := val.FieldByName(k)

        if !field.IsValid() {
            errs = multierror.Append(errs, &MissingFieldError{k})
            continue
        }
    }

    return errs.ErrorOrNil()
}
