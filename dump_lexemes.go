package cptest

import (
    "strings"
	"github.com/logrusorgru/aurora"
)

const (
    AltLineFeed = "\\n"
)

func DumpLexemes(xms []RichText, color aurora.Color) string {
	var str strings.Builder

    x := 0
    for _, xm := range xms {
        if x != 0 && xm.Str != "\n" {
            str.WriteRune(' ')
        }

        if xm.Str == "\n" {
            x = -1

            if xm.Colorful() {
                str.WriteString(Au.Colorize(AltLineFeed, color).String())
                str.WriteRune('\n')
                x++
                continue
            }
        }

        str.WriteString(xm.Colorize(color))
        x++
    }

	return str.String()
}

