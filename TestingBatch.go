package cptest

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"time"
)

// InternalError represents an error that occured due to internal failure in
// TestingBatch.
type InternalError bool

func (e InternalError) Error() string {
	return "internal"
}

// InternalErr is a single instance of the InternalError that may be referred
// to in errors.Is.
const InternalErr = InternalError(true)

// Verdict represents a verdict asssigned by the judge.
type Verdict int

// The set of all possible judge verdicts that can be assigned. The
// abbreviatons are due to competitive programming online judges.
// (Except for IE, that stands for Internal Error)
const (
	OK Verdict = iota
	IE
	WA
	RE
	TL
)

// TestStartCallbackFunc represents a function to be called before a test
// case will be launched. It accepts the id of the test case.
type TestStartCallbackFunc func(int)

// TestStartStub is a stub for TestStartCallback that does nothing.
func TestStartStub(id int) {}

// TestEndCallbackFunc represents a function to be called when a test
// case finishes execution. Usually, one would like to print the test case
// result information.
//
// It accepts the pointer to the TestingBatch that contains verdict, time,
// program's error and output info. it also accepts the Test and the id of
// the Test.
type TestEndCallbackFunc func(*TestingBatch, Test, int)

// TestEndStub is a stub for TestEndCallback that does nothing.
func TestEndStub(b *TestingBatch, test Test, id int) {}

// TestResult carries an output of the process given the ID-th test input
// and an error if any.
type TestResult struct {
    ID int
    Err error
    Out string
}

// TestingBatch is responsible for running tests and evaluating the verdicts
// for tests. For each test case, the verdict and execution time are stored.
// It utilizer an instance of Processer to run tests, and an instance of
// Stopwatcher to track time limit. Optionally, user can set ResultPrinter
// to a custom function to output useful statistics about test case's result.
type TestingBatch struct {
	inputs Inputs

	complete chan TestResult

	Errs map[int]error
	Outs map[int]string
	Diff LexComparison

	Verdicts map[int]Verdict
	Times    map[int]time.Duration

	Proc   Processer
	Swatch Stopwatcher

	TestStartCallback TestStartCallbackFunc
	TestEndCallback   TestEndCallbackFunc
}

// NewTestingBatch will initialize channels and maps inside TestingBatch and
// will assign respective dependency injections.
func NewTestingBatch(inputs Inputs, proc Processer, swatch Stopwatcher) *TestingBatch {
	return &TestingBatch{
		inputs: inputs,

		complete: make(chan TestResult),
		Errs:     make(map[int]error),
		Outs:     make(map[int]string),

		Verdicts: make(map[int]Verdict),
		Times:    make(map[int]time.Duration),

		Proc:   proc,
		Swatch: swatch,

		TestStartCallback: TestStartStub,
		TestEndCallback:   TestEndStub,
	}
}

func (b *TestingBatch) launchTest(id int, in string) {
    defer func() {
        if e := recover(); e != nil {
            b.complete <- TestResult{
                ID: id,
                Err: fmt.Errorf("%w: %v", InternalErr, e),
                Out: "",
            }
        }
    }()

    buf := &bytes.Buffer{}
    err := b.Proc.Run(strings.NewReader(in), buf)

    b.complete <- TestResult{
        ID: id,
        Err: err,
        Out: buf.String(),
    }
}

// Run will lauch test cases in parallel and then will wait for each test to
// finish or for the time to timelimit. When a test is finished the verdict
// and the time it took to execute are remembered. Additionally, ResultPrinter
// is called on the test case's statistics. When a time limit is reached,
// each not-yet-judged test is assigned TL verdict and the ResultPrinter is
// also called on each test.
func (b *TestingBatch) Run() {
	for i, test := range b.inputs.Tests {
		b.TestStartCallback(i + 1)

		go b.launchTest(i+1, test.Input)
	}

	for range b.inputs.Tests {
		var result TestResult

		select {
		case tl := <-b.Swatch.TimeLimit():
            for id := range b.inputs.Tests {
				if _, finished := b.Verdicts[id+1]; !finished {
					b.Verdicts[id+1] = TL
					b.Times[id+1] = tl

					b.TestEndCallback(b, b.inputs.Tests[id], id+1)
				}
			}

			return
		case result = <-b.complete:
		}
        
        id := result.ID
		test := b.inputs.Tests[id-1]

        b.Errs[id] = result.Err
        b.Outs[id] = result.Out
		b.Times[id] = b.Swatch.Elapsed()

		if err := b.Errs[id]; err != nil {
			if errors.Is(err, InternalErr) {
				b.Verdicts[id] = IE
			} else {
				b.Verdicts[id] = RE
			}
		} else {
			lexer := Lexer{}

			got := lexer.Scan(b.Outs[id])
			want := lexer.Scan(test.Output)

			var same bool
			b.Diff, same = lexer.Compare(got, want)

			if !same {
				b.Verdicts[id] = WA
			} else {
				b.Verdicts[id] = OK
			}
		}

		b.TestEndCallback(b, test, id)
	}
}
