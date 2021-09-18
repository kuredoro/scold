package cptest

import (
    //"fmt"
    "reflect"
)

const(
    // NotStructLike is issued when a destination type doesn't behave like
    // a struct. struct and *struct types have the same syntax for manipulating
    // them, so they are considered struct-like.
    NotAStructLike = StringError("not a struct-like")
)

type KVMap = map[string]string

func KVMapUnmarshal(kvm KVMap, data interface{}) error {
    val := reflect.ValueOf(data)

    if val.Kind() == reflect.Ptr {
        val = val.Elem()
    }

    if val.Kind() != reflect.Struct {
        return NotAStructLike
    }

    return nil
}
