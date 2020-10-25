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
    return nil
}

func (b *TestingBatch) Run() {

}
