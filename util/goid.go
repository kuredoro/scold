package util

import (
    "runtime"
    "strings"
    "strconv"
    "fmt"
)

// Borrowed from: https://gist.github.com/metafeather/3615b23097836bc36579100dac376906
func Goid() int {
	var buf [64]byte
	n := runtime.Stack(buf[:], false)
	idField := strings.Fields(string(buf[:n]))[1]
	id, err := strconv.Atoi(idField)
	if err != nil {
		panic(fmt.Sprintf("cannot get gorountine id: %v", err))
	}

	return id
}

