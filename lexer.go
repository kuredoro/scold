package cptest

import (
	"bufio"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

var VALID_INT_MAX_LEN = 10

type LexemeType int

const (
	STRXM LexemeType = iota
	FLOATXM
	INTXM
	FINALXM
)

func IsIntLexeme(xm string) bool {
	_, err := strconv.Atoi(xm)

	return err == nil && len(xm) <= VALID_INT_MAX_LEN
}

func IsFloatLexeme(xm string) bool {
	if xm[0] == '+' || xm[0] == '-' {
		xm = xm[1:]
	}

	parts := strings.Split(xm, ".")

	// 123.456.789 and others
	if len(parts) > 2 {
		return false
	}

	for _, r := range xm {
		if !('0' <= r && r <= '9') && r != '.' {
			return false
		}
	}

	return xm != "."
}

var TypeCheckers = []func(string) bool{
	IsFloatLexeme,
	IsIntLexeme,
}

var MaskGenerators = map[LexemeType]func(*Lexer, string, string) []bool{
	STRXM:   (*Lexer).GenMaskForString,
	FLOATXM: (*Lexer).GenMaskForFloat,
	INTXM:   (*Lexer).GenMaskForInt,
}

// IDEA: Add map[string]interface{} for custom configs from outside of library.
type Lexer struct {
	Precision uint
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

func (l *Lexer) Compare(target, source []string) (rts []RichText, ok bool) {
	rts = make([]RichText, len(target))
	ok = true

	commonLen := len(target)
	if len(source) < commonLen {
		commonLen = len(source)
	}

	for i, xm := range target[:commonLen] {
		rts[i].Str = xm

		targetType := DeduceLexemeType(xm)
		sourceType := DeduceLexemeType(source[i])

		commonType := targetType
		if sourceType < commonType {
			commonType = sourceType
		}

		rts[i].Mask = MaskGenerators[commonType](l, xm, source[i])

		maskEmpty := true
		for _, bit := range rts[i].Mask {
			if bit == true {
				maskEmpty = false
				break
			}
		}

		if !maskEmpty {
			ok = false
		}
	}

	for i := commonLen; i < len(target); i++ {
		rts[i].Str = target[i]

		rts[i].Mask = make([]bool, len(target[i]))
		for mi := range rts[i].Mask {
			rts[i].Mask[mi] = true
		}

		ok = false
	}

	return
}

func DeduceLexemeType(xm string) LexemeType {
	for i := int(STRXM) + 1; i != int(FINALXM); i++ {
		// As any lexeme *is* a string, the function IsStringLexeme is omitted.
		if !TypeCheckers[i-1](xm) {
			return LexemeType(i - 1)
		}
	}

	return LexemeType(FINALXM - 1)
}

func (l *Lexer) GenMaskForString(target, source string) (mask []bool) {
	commonLen := len(target)
	if len(source) < commonLen {
		commonLen = len(source)
	}

	mask = make([]bool, len(target))

	for i := 0; i < commonLen; i++ {
		mask[i] = target[i] != source[i]
	}

	for i := commonLen; i < len(target); i++ {
		mask[i] = true
	}

	return
}

func (l *Lexer) GenMaskForInt(target, source string) (mask []bool) {
	mask = make([]bool, len(target))

	if target == "" || source == "" {
		return
	}

	if target[0] == '-' && source[0] != '-' || target[0] == '+' && source[0] == '-' {
		mask[0] = true
	}

	targetVal, _ := strconv.Atoi(target)
	if targetVal < 0 {
		targetVal = -targetVal
	}

	sourceVal, _ := strconv.Atoi(source)
	if sourceVal < 0 {
		sourceVal = -sourceVal
	}

	if targetVal != sourceVal {
		for i := range mask {
			mask[i] = true
		}
	}

	return
}

func (l *Lexer) GenMaskForFloat(target, source string) (mask []bool) {
	targetWhole := strings.Split(target, ".")[0]

	sourceWhole := strings.Split(source, ".")[0]
	if sourceWhole == "" {
		sourceWhole = "0"
	}

	mask = l.GenMaskForInt(targetWhole, sourceWhole)

	// dot is never colored
	mask = append(mask, false)

	tragetFracStart := strings.IndexRune(target, '.') + 1
	if tragetFracStart == 0 {
		tragetFracStart = len(target)
	}

	sourceFracStart := strings.IndexRune(source, '.') + 1
	if sourceFracStart == 0 {
		sourceFracStart = len(source)
	}

	targetFrac := target[tragetFracStart:]
	sourceFrac := source[sourceFracStart:]

	if len(targetFrac) > len(sourceFrac) {
		sourceFrac += strings.Repeat("0", len(targetFrac)-len(sourceFrac))
	}

	fracMask := make([]bool, len(targetFrac))
	equal := true

	for i := 0; i < len(targetFrac); i++ {
		if targetFrac[i] != sourceFrac[i] {
			equal = false
		}

		if !equal && i < int(l.Precision) {
			fracMask[i] = true
		}
	}

	mask = append(mask, fracMask...)

	return
}
