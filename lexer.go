package cptest

import (
	"bufio"
	"strings"
	"unicode"
	"unicode/utf8"
)

type LexComparison struct {
	Got  []RichText
	Want []RichText
}

type Lexer struct{}

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
				return i + width, data[start : i+width], nil
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
func (l *Lexer) Scan(text string) (xms []string) {
	r := strings.NewReader(text)
	s := bufio.NewScanner(r)
	s.Split(ScanLexemes)

	for s.Scan() {
		xms = append(xms, s.Text())
	}

	return
}

func (l *Lexer) Compare(got, want []string) (diff LexComparison, ok bool) {
	ok = len(got) == len(want)

	// Make got always smaller or equal want to simplify code.
	swapped := false
	if len(got) > len(want) {
		tmp := got
		got = want
		want = tmp
		swapped = true
	}

	for i := range got {
        gotRt := RichText{
            Str: got[i],
            Mask: make([]bool, len(got[i])),
        }

        wantRt := RichText{
            Str: want[i],
            Mask: make([]bool, len(want[i])),
        }

        commonSize := len(got[i])
        if len(got[i]) > len(want[i]) {
            commonSize = len(want[i])
        }

        for j := 0; j < commonSize; j++ {
            if got[i][j] != want[i][j] {
                gotRt.Mask[j] = true
                wantRt.Mask[j] = true
            }
        }

        for j := commonSize; j < len(got[i]); j++ {
            gotRt.Mask[j] = true
        }

        for j := commonSize; j < len(want[i]); j++ {
            wantRt.Mask[j] = true
        }

        if gotRt.Colorful() || wantRt.Colorful() {
            ok = false
        }

		diff.Got = append(diff.Got, gotRt)
		diff.Want = append(diff.Want, wantRt)
	}

	for i := len(got); i < len(want); i++ {
        mask := make([]bool, len(want[i]))
        for j := range want[i] {
            mask[j] = true
        }

		diff.Want = append(diff.Want, RichText{
			want[i],
            mask,
		})
	}

	if swapped {
		tmp := diff.Got
		diff.Got = diff.Want
		diff.Want = tmp
	}

	return
}
