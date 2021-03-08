package cptest

import (
	"io"
	"sync"
)

// ProcessResult contains the text printed to stdout and stderr by the process
// and the exit code returned upon termination.
type ProcessResult struct {
    ExitCode int
    Stdout string
    Stderr string
}

// Processer interface abstracts away the concept of the executable under
// testing.
type Processer interface {
	Run(io.Reader) (ProcessResult, error)
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
func (p *SpyProcesser) Run(r io.Reader) (ProcessResult, error) {
	p.mu.Lock()
	p.callCount++
	p.mu.Unlock()
	return p.Proc.Run(r)
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
type ProcesserFunc func(io.Reader) (ProcessResult, error)

// Run will call the underlying Go function to compute the result.
func (p ProcesserFunc) Run(r io.Reader) (ProcessResult, error) {
	return p(r)
}
