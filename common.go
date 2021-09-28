package cptest

import (
	"fmt"
)

// StringError is an error type whose values can be constant and compared against deterministically with == operator. An error type that solves the problems of sentinel errors.
type StringError string

func (e StringError) Error() string {
	return string(e)
}

// LineRangeError appends line information to the error message. TODO: remove -> It is mainly used
// to test that errors are produced for correct lines.
type LineRangeError struct {
	Begin int
	End   int
	Err   error
}

func (e *LineRangeError) Error() string {
	if e.Begin+1 == e.End {
		return fmt.Sprintf("line %d: %v", e.Begin, e.Err)
	}

	return fmt.Sprintf("lines %d-%d: %v", e.Begin, e.End-1, e.Err)
}

func (e *LineRangeError) Unwrap() error {
	return e.Err
}

type TestError struct {
	TestNum int
	Err     error
}

func (e *TestError) Error() string {
	return fmt.Sprintf("test %d: %v", e.TestNum, e.Err)
}

func (e *TestError) Unwrap() error {
	return e.Err
}
