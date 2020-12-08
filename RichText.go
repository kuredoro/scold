package cptest

import (
	"sort"
	"strings"

	"github.com/logrusorgru/aurora"
)

type RichText struct {
	Str       string
	Partition []int
}

func (rt *RichText) Colorize(color aurora.Color) string {
	var str strings.Builder

	sort.Ints(rt.Partition)

	start := 0
	for i, end := range rt.Partition {
		part := rt.Str[start:end]

		if i&1 == 0 {
			str.WriteString(part)
		} else {
			str.WriteString(aurora.Colorize(part, color).String())
		}

		start = end
	}

	return str.String()
}
