package cptest

import (
	"fmt"
)

// StringError is an error type whose values can be constant and compared against deterministically with == operator. An error type that solves the problems of sentinel errors.
type StringError string

func (e StringError) Error() string {
	return string(e)
}

// LineError appends line information to the error message. It is mainly used
// to test that errors are produced for correct lines.
type LineError struct {
	Line int
	Err  error
}

func (e *LineError) Error() string {
	return fmt.Sprintf("line %d: %v", e.Line, e.Err)
}

func (e *LineError) Unwrap() error {
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
