package scold

import (
	"fmt"
	"reflect"
	"strings"
)

// StringError is an error type whose values can be constant and compared
// against deterministically with == operator. An error type that solves the
// problems of sentinel errors.
type StringError string

// Error makes StringError satisfy error interface.
func (e StringError) Error() string {
	return string(e)
}

// StringWarning is same as StringError except it allows users to differentiate
// warnings using errors.As
type StringWarning string

// Error prints the warning message.
func (w StringWarning) Error() string {
    return string(w)
}

// LineRangeError is used to amend information about the location of the error
// within the source code. The source code lines in
// question are stored in Lines (for printing purposes).
type LineRangeError struct {
	Begin int
	Lines []string
	Err   error
}

// Error renders contents of Err.Error() preceeded by a line
// number (or a line range), followed by the line numbered code snippet (if provided)
// on a new line. The error string is terminated with '\n', unless Lines is not
// nil. It uses Reason and CodeSnippet under the hood.
func (e *LineRangeError) Error() string {
	if len(e.Lines) == 0 {
		return e.Reason()
	}

	var msg strings.Builder

	msg.WriteString(e.Reason())
	msg.WriteByte('\n')
	msg.WriteString(e.CodeSnippet())

	return msg.String()
}

// Reason is a stripped down version of Error in sense that it returns only
// the underlying error's message and a line information, with no '\n' at the
// end.
func (e *LineRangeError) Reason() string {
	if len(e.Lines) == 0 || len(e.Lines) == 1 {
		return fmt.Sprintf("line %d: %v", e.Begin, e.Err)
	}

	return fmt.Sprintf("lines %d-%d: %v", e.Begin, e.Begin+len(e.Lines), e.Err)
}

// CodeSnippet is responsible for producing pretty printed source code lines
// with corresponding line numbers to the left of them.
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

// Unwrap makes LineRangeError usable with built-in errors package.
func (e *LineRangeError) Unwrap() error {
	return e.Err
}

// TestError is a scold centric error type that maps a test's index inside
// inputs.txt to a particular error Err.
type TestError struct {
	TestNum int
	Err     error
}

// Error renders the underlying error preceeded with corresponding test number.
func (e *TestError) Error() string {
	return fmt.Sprintf("test %d: %v", e.TestNum, e.Err)
}

// Unwrap makes TestError usable with built-in errors package.
func (e *TestError) Unwrap() error {
	return e.Err
}

// FieldError is a generic error that can be produced while unmarshaling string
// maps. It enriches error Err with the name of the field relevant to it.
type FieldError struct {
	FieldName string
	Err       error
}

// Error renders the underlying error preceeded with the field's
// name.
func (e *FieldError) Error() string {
	return fmt.Sprintf("field %q: %v", e.FieldName, e.Err)
}

// Unwrap makes FieldError usable with built-in errors package.
func (e *FieldError) Unwrap() error {
	return e.Err
}

// NotValueOfTypeError is a generic basic error type that describes a value
// and a type that don't match. This error is produced mainly during
// unmarshaling string maps.
type NotValueOfTypeError struct {
	Type  string
	Value string
    Err error
}

// Error prints a human-readable message describing what value doesn't match
// what type.
func (e *NotValueOfTypeError) Error() string {
	return fmt.Sprintf("value %q doesn't match %v type", e.Value, e.Type)
}

// Unwrap returns the reason for the error, if available.
func (e *NotValueOfTypeError) Unwrap() error {
    return e.Err
}

// Equal is used to define equality on the pointers of NotValueOfTypeError.
// Used by go-testdeep.
func (e *NotValueOfTypeError) Equal(other *NotValueOfTypeError) bool {
	return e.Type == other.Type && e.Value == other.Value
}

// NotTextUnmarshalableTypeError is a panic error. Value of this type is
// passed to panic(), mainly during unmarshaling of string maps.
type NotTextUnmarshalableTypeError struct {
	Field    string
	Type     reflect.Kind
	TypeName string
}

// Error renders a helpful message addressed to the developer, describing
// possible reasons as to why type could not be unmarshaled.
func (e *NotTextUnmarshalableTypeError) Error() string {
	return fmt.Sprintf("field %q is of type %v (%v) and cannot be unmarshaled from string, because it is not of fundamental type or because the type doesn't implement encoding.TextUnmarshaler interface", e.Field, e.TypeName, e.Type)
}

// Equal is used to define equality on NotTextUnmarshalableTypeError pointers.
// Used by go-testdeep.
func (e *NotTextUnmarshalableTypeError) Equal(other *NotTextUnmarshalableTypeError) bool {
	return e.Field == other.Field && e.Type == other.Type && e.TypeName == other.TypeName
}
