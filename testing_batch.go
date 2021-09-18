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

const TLError StringError = "Time limit exceeded"

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

	Verdicts   map[int]Verdict
	Times      map[int]time.Duration
	startTimes map[int]time.Time

	Proc          Processer
	procCancels   map[int]func()
	procCancelsMu sync.Mutex

	ThreadPool WorkerPool

	Swatch Stopwatcher

	TestStartCallback TestStartCallbackFunc
	TestEndCallback   TestEndCallbackFunc
}

// NewTestingBatch will initialize channels and maps inside TestingBatch and
// will assign respective dependency injections.
func NewTestingBatch(inputs Inputs, proc Processer, swatch Stopwatcher, pool WorkerPool) *TestingBatch {
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

		Verdicts:   make(map[int]Verdict),
		Times:      make(map[int]time.Duration),
		startTimes: make(map[int]time.Time),

		Proc:        proc,
		procCancels: make(map[int]func()),

		ThreadPool: pool,

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
	if _, exists := b.procCancels[id]; !exists {
		b.procCancels[id] = cancel
	} else {
		b.procCancelsMu.Unlock()
		cancel()
		b.complete <- TestResult{
			ID:  id,
			Err: TLError,
		}
		return
	}
	b.procCancelsMu.Unlock()

	out, err := b.Proc.Run(ctx, strings.NewReader(in))

	if ctx.Err() != nil {
		err = TLError
	}

	b.complete <- TestResult{
		ID:  id,
		Err: err,
		Out: out,
	}
}

func (b *TestingBatch) nextOldestRunning(previous int) int {
	for id := previous + 1; id < len(b.inputs.Tests)+1; id++ {
		_, finished := b.Verdicts[id]
		if !finished {
			return id
		}
	}

	return len(b.inputs.Tests) + 1
}

// Run will lauch test cases in parallel and then will wait for each test to
// finish or for the time to timelimit. When a test is finished the verdict
// and the time it took to execute are remembered. Additionally, ResultPrinter
// is called on the test case's statistics. When a time limit is reached,
// each not-yet-judged test is assigned TL verdict and the ResultPrinter is
// also called on each test.
func (b *TestingBatch) Run() {
	nextTestID := 1
	for ; nextTestID-1 < len(b.inputs.Tests) && nextTestID-1 < b.ThreadPool.WorkerCount(); nextTestID++ {
		// Local variable is deliberate, since RunnableFunc below will capture
		// variables by reference, nextTestID will be len(b.inputs.Tests)+1 when
		// the worker picks up the job, and so cause panic
		id := nextTestID
		err := b.ThreadPool.Execute(RunnableFunc(func() {
			b.launchTest(id, b.inputs.Tests[id-1].Input)
		}))

		if err != nil {
			break
		}

		b.TestStartCallback(id)
		b.startTimes[id] = b.Swatch.Now()
	}

	oldestRunningID := 1
	for len(b.Verdicts) != len(b.inputs.Tests) {
		var result TestResult

		select {
		case <-b.Swatch.TimeLimit(b.startTimes[oldestRunningID]):
			b.procCancelsMu.Lock()

			if cancel, exists := b.procCancels[oldestRunningID]; exists {
				cancel()
			} else {
				// Notify the launchTest func not to run the thread
				b.procCancels[oldestRunningID] = func() {}
			}

			b.procCancelsMu.Unlock()

			oldestRunningID = b.nextOldestRunning(oldestRunningID)
			continue
		case result = <-b.complete:
		}

		// A worker is now free, run another test if any
		if nextTestID-1 < len(b.inputs.Tests) {
			id := nextTestID
			err := b.ThreadPool.Execute(RunnableFunc(func() {
				b.launchTest(id, b.inputs.Tests[id-1].Input)
			}))

			if err == nil {
				b.TestStartCallback(id)
				b.startTimes[id] = b.Swatch.Now()
				nextTestID++
			}
		}

		id := result.ID
		test := b.inputs.Tests[id-1]

		b.Errs[id] = result.Err
		b.Outs[id] = result.Out
		b.Times[id] = b.Swatch.Elapsed(b.startTimes[id])

		answerLexemes := b.Lx.Scan(test.Output)
		b.RichAnswers[id], _ = b.Lx.Compare(answerLexemes, nil)

		if result.Err == TLError {
			b.Verdicts[id] = TL
		} else if result.Err != nil {
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
