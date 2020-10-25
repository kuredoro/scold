package cptest

type Verdict int

const (
    OK Verdict = iota
    IE
    WA
    RE
)

type TestingBatch struct {
    inputs Inputs
    proc Processer

    Stat map[int]Verdict
}

func NewTestingBatch(inputs Inputs, proc Processer) *TestingBatch {
    return &TestingBatch{
        inputs: inputs,
        proc: proc,
        Stat: make(map[int]Verdict),
    }
}

func (b *TestingBatch) Run() {
    for i, test := range b.inputs.Tests {

        err := b.proc.Run(i + 1, test.Input)
        if err != nil {
            b.Stat[i + 1] = IE
            continue
        }

        id := b.proc.WaitCompleted()

        if b.proc.GetError(id) != nil {
            b.Stat[i + 1] = RE
            continue
        }

        if test.Output != b.proc.GetOutput(i + 1) {
            b.Stat[i + 1] = WA
            continue
        }

        b.Stat[i + 1] = OK
    }
}
