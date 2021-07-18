package main

import (
    "fmt"
    "strings"
)

type ProgressBar struct {
    Total int
    Current int
    Width int
    Header string
}

func (b *ProgressBar) String() string {
    arrowPos := b.Width * b.Current / b.Total

    var bar strings.Builder
    bar.WriteString(strings.Repeat("=", arrowPos))
    if arrowPos != b.Width {
        bar.WriteByte('>')
        bar.WriteString(strings.Repeat(" ", b.Width-arrowPos-1))
    }

    barStr := []byte(bar.String())
    barStr[0] = '['
    barStr[len(barStr)-1] = ']'

    return fmt.Sprintf("%s %s %d/%d", b.Header, string(barStr), b.Current, b.Total)
}
