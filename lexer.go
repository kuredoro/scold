package cptest

import (
	"bufio"
	"strings"
    "unicode"
    "unicode/utf8"
)

type LexSequence []string

type Lexer struct {

}

// ScanLexemes is a split function for bufio.Scanner. It is same as
// bufio.ScanWords, except that it treats \n character in a special way.
// \n cannot be in any lexeme, except for "\n" itself. Hence, several
// \n\n are parsed as separate lexemes ("\n", "\n").
// It will never return an empty lexeme.
// The definition of other spaces is set by unicode.IsSpace.
func ScanLexemes(data []byte, atEOF bool) (advance int, token []byte, err error) {
	// Skip leading spaces.
	start := 0
	for width := 0; start < len(data); start += width {
		var r rune
		r, width = utf8.DecodeRune(data[start:])
		if r == '\n' || !unicode.IsSpace(r) {
			break
		}
	}

	// Scan until space, marking end of word.
	for width, i := 0, start; i < len(data); i += width {
		var r rune
		r, width = utf8.DecodeRune(data[i:])

        if r == '\n' {
            if i == start {
                return i + width, data[start:i + width], nil
            }

            return i, data[start:i], nil
        }

		if unicode.IsSpace(r) {
			return i + width, data[start:i], nil
		}
	}

	// If we're at EOF, we have a final, non-empty, non-terminated word. Return it.
	if atEOF && len(data) > start {
		return len(data), data[start:], nil
	}

	// Request more data.
	return start, nil, nil
}

// Scan will break the text into lexemes and return them. A lexeme
// is either a string consisting of not unicode.IsSpace characters,
// or a single newline character.
// The returned LexSequence is never nil.
func (l *Lexer) Scan(text string) (seq LexSequence) {
    seq = LexSequence{}

    r := strings.NewReader(text)
    s := bufio.NewScanner(r)
    s.Split(ScanLexemes)

    for s.Scan() {
        seq = append(seq, s.Text())
    }

    return
}
