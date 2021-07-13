package cptest

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"
)

// DefaultPrecision is the value the TestingBatch lexer's precision is
// initialized with in NewTestingBatch function.
var DefaultPrecision uint = 6

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
	ID  int
	Err error
	Out ProcessResult
}

// TestingBatch is responsible for running tests and evaluating the verdicts
// for tests. For each test case, the verdict and execution time are stored.
// It utilizer an instance of Processer to run tests, and an instance of
// Stopwatcher to track time limit. Optionally, user can set ResultPrinter
// to a custom function to output useful statistics about test case's result.
type TestingBatch struct {
	inputs Inputs

	complete chan TestResult

	Errs        map[int]error
	Outs        map[int]ProcessResult
	RichOuts    map[int][]RichText
	RichAnswers map[int][]RichText
	Lx          *Lexer

	Verdicts map[int]Verdict
	Times    map[int]time.Duration

	Proc   Processer
    procCancels map[int]func()
    procCancelsMu sync.Mutex

    ThreadPool WorkerPool

	Swatch Stopwatcher

	TestStartCallback TestStartCallbackFunc
	TestEndCallback   TestEndCallbackFunc
}

// NewTestingBatch will initialize channels and maps inside TestingBatch and
// will assign respective dependency injections.
func NewTestingBatch(inputs Inputs, proc Processer, swatch Stopwatcher) *TestingBatch {
	precision, err := strconv.Atoi(inputs.Config["prec"])
	if err != nil {
		precision = int(DefaultPrecision)
	}

	return &TestingBatch{
		inputs: inputs,

		complete:    make(chan TestResult),
		Errs:        make(map[int]error),
		Outs:        make(map[int]ProcessResult),
		RichOuts:    make(map[int][]RichText),
		RichAnswers: make(map[int][]RichText),

		Lx: &Lexer{
			Precision: uint(precision),
		},

		Verdicts: make(map[int]Verdict),
		Times:    make(map[int]time.Duration),

		Proc:   proc,
        procCancels: make(map[int]func()),

		Swatch: swatch,

		TestStartCallback: TestStartStub,
		TestEndCallback:   TestEndStub,
	}
}

func (b *TestingBatch) launchTest(id int, in string) {
	defer func() {
		if e := recover(); e != nil {
			b.complete <- TestResult{
				ID:  id,
				Err: fmt.Errorf("internal: %v", e),
				Out: ProcessResult{},
			}
		}
	}()

    ctx, cancel := context.WithCancel(context.Background())

    b.procCancelsMu.Lock()
    b.procCancels[id] = cancel
    b.procCancelsMu.Unlock()

	out, err := b.Proc.Run(ctx, strings.NewReader(in))

	b.complete <- TestResult{
		ID:  id,
		Err: err,
		Out: out,
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
            tled := make([]int, 0, len(b.inputs.Tests))
			for id := range b.inputs.Tests {
				if _, finished := b.Verdicts[id+1]; !finished {
                    tled = append(tled, id+1)
				}
			}

            for _, id := range tled {
                b.Verdicts[id] = TL
                b.Times[id] = tl

                answerLexemes := b.Lx.Scan(b.inputs.Tests[id-1].Output)

                rich := make([]RichText, len(answerLexemes))
                for i, xm := range answerLexemes {
                    rich[i] = RichText{xm, make([]bool, len(xm))}
                }

                b.RichAnswers[id] = rich

                b.procCancelsMu.Lock()
                b.procCancels[id]()
                b.procCancelsMu.Unlock()

                b.TestEndCallback(b, b.inputs.Tests[id-1], id)
            }

            for range tled {
                <-b.complete
            }

			return
		case result = <-b.complete:
		}

		id := result.ID
		test := b.inputs.Tests[id-1]

		b.Errs[id] = result.Err
		b.Outs[id] = result.Out
		b.Times[id] = b.Swatch.Elapsed()

		answerLexemes := b.Lx.Scan(test.Output)
		b.RichAnswers[id], _ = b.Lx.Compare(answerLexemes, nil)

		if err := b.Errs[id]; err != nil {
			b.Verdicts[id] = IE
		} else if b.Outs[id].ExitCode != 0 {
			b.Verdicts[id] = RE
		} else {
			got := b.Lx.Scan(b.Outs[id].Stdout)

			var okOut, okAns bool
			b.RichOuts[id], okOut = b.Lx.Compare(got, answerLexemes)
			b.RichAnswers[id], okAns = b.Lx.Compare(answerLexemes, got)

			same := okOut && okAns

			if !same {
				b.Verdicts[id] = WA
			} else {
				b.Verdicts[id] = OK
			}
		}

		b.TestEndCallback(b, test, id)
	}
}
