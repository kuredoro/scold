package cptest

import (
	"github.com/logrusorgru/aurora"
	"strings"
)

// AltLineFeed is the representation of LF in textual form, that replaces LF
// when it's colorized.
const AltLineFeed = "\\n"

// DumpLexemes is used to transform array of possibly colorized lexemes into a
// human readable format. The lexemes are separated by spaces. There are no
// trailing spaces. Colorized newlines are replaced by printable AltLineFeed
// string + a newline.
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
