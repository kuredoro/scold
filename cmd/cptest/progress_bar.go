package main

import (
    "fmt"
)

type ProgressBar struct {
    Total int
    Current int
    Header string
}

func (b *ProgressBar) String() string {
    return fmt.Sprintf("%s %d/%d", b.Header, b.Current, b.Total)
}
