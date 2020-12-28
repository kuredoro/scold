package cptest

import (
	"io"
	"sync"
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

	mu        sync.Mutex
	callCount int
}

// Run will execute the Run function of the inner processer, but will
// also increase the call count by one.
func (p *SpyProcesser) Run(r io.Reader, w io.Writer) error {
	p.mu.Lock()
	p.callCount++
	p.mu.Unlock()
	return p.Proc.Run(r, w)
}

// CallCount will return the number of times Run was called. Can be called
// concurrently.
func (p *SpyProcesser) CallCount() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.callCount
}

// ProcesserFunc represents an implementation of Processer that instead
// of a real OS-level process executes Go code.
type ProcesserFunc func(io.Reader, io.Writer) error

// Run will call the underlying Go function to compute the result.
func (p ProcesserFunc) Run(r io.Reader, w io.Writer) error {
	return p(r, w)
}
