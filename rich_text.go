package scold

import (
	"strings"

	"github.com/logrusorgru/aurora"
)

// Au is used to colorize output of several functions. A user of the library
// can change its value to disable colored output. Refer to the aurora
// readme for that.
var Au aurora.Aurora

func init() {
	Au = aurora.NewAurora(true)
}

// RichText represents a text data with additional color metadata in a form
// of a bitmask. The characters may be either colored or uncolored. The _color_
// might represent a literal color or a formatting style like bold or italics.
type RichText struct {
	Str  string
	Mask []bool
}

// Colorful returns whether at least one character in Str has color.
func (rt RichText) Colorful() bool {
	for _, v := range rt.Mask {
		if v {
			return true
		}
	}

	return false
}

// Colorize returns Str with ASCII escape codes actually
// embedded inside it to enable colors. The resulting string then
// can be printed on the screen and it'll be colorful, for example.
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
