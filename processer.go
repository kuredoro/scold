package cptest

import (
	"io"
)

// Processer interface abstracts away the concept of the executable under
// testing.
type Processer interface {
	Run(io.Reader, io.Writer) error
}

// SpyProcesser is a test double that proxies another processer.
// It additionally stores the number of calls made to the Run function.
type SpyProcesser struct {
	Proc Processer

	CallCount int
}

// Run will execute the Run function of the inner processer, but will
// also increase the call count by one.
func (p *SpyProcesser) Run(r io.Reader, w io.Writer) error {
	p.CallCount++
	return p.Proc.Run(r, w)
}

// ProcesserFunc represents an implementation of Processer that instead
// of a real OS-level process executes Go code.
type ProcesserFunc func(io.Reader, io.Writer) error

// Run will call the underlying Go function to compute the result.
func (p ProcesserFunc) Run(r io.Reader, w io.Writer) error {
	return p(r, w)
}
