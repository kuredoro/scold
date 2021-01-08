package cptest

import (
	"strings"

	"github.com/logrusorgru/aurora"
)

// Au is used to colorize output of several functions. A user of the library
// can change the value of it to disable colored output. Refer to the aurora
// documentation for that.
var Au aurora.Aurora

func init() {
	Au = aurora.NewAurora(true)
}

type RichText struct {
	Str  string
	Mask []bool
}

func (rt RichText) Colorful() bool {
	for _, v := range rt.Mask {
		if v {
			return true
		}
	}

	return false
}

func (rt RichText) Colorize(color aurora.Color) string {
	var str strings.Builder

	start := 0
	for start != len(rt.Mask) {
		end := start + 1
		for ; end != len(rt.Mask); end++ {
			if rt.Mask[start] != rt.Mask[end] {
				break
			}
		}

		part := rt.Str[start:end]
		if rt.Mask[start] {
			str.WriteString(Au.Colorize(part, color).String())
		} else {
			str.WriteString(part)
		}

		start = end
	}

	return str.String()
}
