package cptest

import (
	"fmt"
	"strings"
    "reflect"
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
	Lines []string
	Err   error
}

func (e *LineRangeError) Error() string {
	if len(e.Lines) == 0 {
		return fmt.Sprintf("line %d: %v", e.Begin, e.Err)
	}

	var msg strings.Builder
	if len(e.Lines) == 1 {
		msg.WriteString(fmt.Sprintf("line %d: %v\n", e.Begin, e.Err))
	} else {
		msg.WriteString(fmt.Sprintf("lines %d-%d: %v\n", e.Begin, e.Begin+len(e.Lines), e.Err))
	}

	msg.WriteString(e.CodeSnippet())

	return msg.String()
}

func (e *LineRangeError) CodeSnippet() string {
	var msg strings.Builder

	for i, line := range e.Lines {
		if line != "" && line[len(line)-1] == '\n' {
			line = line[:len(line)-1]
		}

		msg.WriteString(fmt.Sprintf("% 6d| %s\n", e.Begin+i, line))
	}

	return msg.String()
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
	return fmt.Sprintf("field %q is of type %v (%v) and cannot be unmarshaled from string, because it is not of fundamental type or because the type doesn't implement encoding.TextUnmarshaler interface", e.Field, e.TypeName, e.Type)
}

func (e *NotStringUnmarshalableTypeError) Equal(other *NotStringUnmarshalableTypeError) bool {
	return e.Field == other.Field && e.Type == other.Type && e.TypeName == other.TypeName
}
