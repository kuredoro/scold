package cptest

import (
	"bufio"
	"fmt"
	"strings"
)

// InputsError represents an error produced while scanning inputs file.
type InputsError string

func (e InputsError) Error() string {
	return string(e)
}

// A set of errors that may be produced during scanning of the inputs file.
// These replace sentinel errors making the errors be comparable with equals
// operator.
const (
	IOSeparatorMissing = InputsError("IO separator missing")
	KeyMissing         = InputsError("key cannot be empty")
)

// LinedError appends line information to the error message. It is mainly used
// to test that errors are produced for correct lines.
type LinedError struct {
	Header string
	Line   int
	Err    error
}

func (e *LinedError) Error() string {
	return fmt.Sprintf("%s: line %d: %v", e.Header, e.Line, e.Err)
}

func (e *LinedError) Unwrap() error {
	return e.Err
}

// The set of delimeters used when partitioning inputs file.
const (
	IODelim   = "---"
	TestDelim = "==="
)

// Test represents a single test case: an input and the expected output.
type Test struct {
	Input  string
	Output string
}

// Inputs contains all information located in the inputs file: tests and
// the set of key-value pairs that were provided.
type Inputs struct {
	Tests  []Test
	Config map[string]string
}

// ScanKeyValuePair parses the key-value pair of form 'key=value'.
// Strings without assignment are treated as keys with empty value.
// Strings with assignment but with empty key are erroneous.
// The space around key and value respectively is trimmed.
func ScanKeyValuePair(line string) (string, string, error) {
	parts := strings.SplitN(line, "=", 2)

	if len(parts) == 1 {
		cleanLine := strings.TrimSpace(line)

		if cleanLine == "" {
			return "", "", nil
		}

		return parts[0], "", nil
	}

	key := strings.TrimSpace(parts[0])
	val := strings.TrimSpace(parts[1])

	if key == "" {
		return "", "", KeyMissing
	}

	return key, val, nil
}

// ScanConfig tries to parse a stream of key-value pairs. It expects each pair
// to be located on a dedicated line. Duplicate keys are allowed, the later
// version is preferred.
func ScanConfig(text string) (m map[string]string, errs []error) {
	s := bufio.NewScanner(strings.NewReader(text))

	m = make(map[string]string)
	for lineNum := 1; s.Scan(); lineNum++ {
		key, val, err := ScanKeyValuePair(s.Text())

		if err != nil {
			errs = append(errs, &LinedError{
				Header: "scan config",
				Line:   lineNum,
				Err:    err,
			})
			continue
		}

		if key == "" && val == "" {
			continue
		}

		m[key] = val
	}

	return
}

// SplitByInlinedPrefixN works in the same way as strings.SplitN. However,
// it does one additional thing. It matches the *prefixes of the lines*
// for equality with the delimeter. Upon match the entire line is discarded.
//
// If text doesn't contain the delimeter, only one part is returned.
// User can specify the number of parts they want at most via the third
// argument.
func SplitByInlinedPrefixN(text, delim string, n int) (parts []string) {

	var str strings.Builder

	s := bufio.NewScanner(strings.NewReader(text))
	for s.Scan() {

		if (n == 0 || len(parts)+1 < n) && strings.HasPrefix(s.Text(), delim) {
			part := str.String()
			parts = append(parts, part)

			str = strings.Builder{}
			continue
		}

		str.WriteString(s.Text())
		str.WriteRune('\n')
	}

	part := str.String()
	parts = append(parts, part)

	return
}

// ScanTest parses a single test case: input and output, separated with the
// Input/Output separator. It also trims space around input and output
// line-wise. If separator is absent, it returns an error.
func ScanTest(testStr string) (Test, []error) {

	if testStr == "" {
		return Test{}, nil
	}

	parts := SplitByInlinedPrefixN(testStr, IODelim, 2)

	if len(parts) == 1 {
		return Test{}, []error{IOSeparatorMissing}
	}

	test := Test{
		Input:  parts[0],
		Output: parts[1],
	}

	return test, nil
}

// ScanInputs is the main routine for parsing inputs file. It splits the input
// by test case separator, and tries to parse each individual test case one by
// one. If the first meaningful test could not be parsed without errors, it is
// interpreted as a configuration and parsed again. There may be multiple
// configurations per file. The later ones overwrite keys set by former ones.
// The empty tests are skipped (those that don't contain input, output and the
// separator). If test case could not be parsed, parsing continues to the next
// test case, but the errors are accumulated and returned together.
func ScanInputs(text string) (inputs Inputs, errs []error) {
	inputs.Config = make(map[string]string)

	parts := SplitByInlinedPrefixN(text, TestDelim, 0)

	testNum := 0
	for partNum, part := range parts {
		test, testErrs := ScanTest(part)

		// Try to parse config
		if testErrs != nil && partNum == 0 {
			config, configErrs := ScanConfig(part)

			if configErrs != nil {
				errs = append(errs, configErrs...)
			}

			for k, v := range config {
				inputs.Config[k] = v
			}

			continue
		}

		// Skip empty tests
		if testErrs == nil && test.Input == "" && test.Output == "" {
			continue
		}

		testNum++

		if testErrs != nil {
			for i, err := range testErrs {
				testErrs[i] = fmt.Errorf("test %d: %w", testNum, err)
			}

			errs = append(errs, testErrs...)
			continue
		}

		inputs.Tests = append(inputs.Tests, test)
	}

	return
}
