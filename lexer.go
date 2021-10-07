package cptest

import (
	"bufio"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

// ValidIntMaxLen is maximum number of digits a lexeme may have to be
// considered an int
var ValidIntMaxLen = 10

type lexemeType int

// Available lexeme types. The following invariant holds:
// each type is a specialization of all types whose numerical value is less
// than that of self. For example, 42 is a float and is an int, but 42.2 is
// a float but not an int. Hence, int is a specialization of float. A type T
// is a specialization of a type U if any value of type T is of type U also.
//
// The consequence is that between any two types T and U from the list
// there's always a specialization relationship, but in gerenal, this is not
// the case. For example: imagine a lexeme type 'hash' that classifies
// strings of form 2400f9b. The float is not a specialization
// of hash, because 42.2 is not a hash, and likewise the
// hash is not a specialization of float, because 2400f9b is not a float.
const (
	STRXM lexemeType = iota
	FLOATXM
	INTXM
	FINALXM
)

// IsIntLexeme returns true if the string represents a signed integer.
// Additionally, it should contain not more than VALID_INT_MAX_LEN digits.
func IsIntLexeme(xm string) bool {
	_, err := strconv.Atoi(xm)

	return err == nil && len(xm) <= ValidIntMaxLen
}

// IsFloatLexeme returns true if the string represents a floating-point value.
// Although, there can be no floating-point inside of it. A floating-point
// value is of form int_part['.' ('0'-'9')*]
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

// TypeCheckers defines a list type checking functions (TCF).
// The type checker for string is omitted because it always returns true.
// Hence, the index of TCF corresponds to a type of numerical value `index+1`.
var TypeCheckers = []func(string) bool{
	IsFloatLexeme,
	IsIntLexeme,
}

// MaskGenerators lists all of the mask generating functions (MGF). MGFs are
// defined only for arguments of the same type. I.e., there's no MGF for float
// and int, only for float/float, and int/int. If differnt types must be
// assessed, the MGF of their common type_ must be called. The common type
// between two types T and U exists if specialization relationship between them
// exists and is the least specialized type. The index of MGF in this array
// corresponds to the numerical value of the type of which MGF's arguments are.
var MaskGenerators = []func(*Lexer, string, string) []bool{
	(*Lexer).GenMaskForString,
	(*Lexer).GenMaskForFloat,
	(*Lexer).GenMaskForInt,
}

// IDEA: Add map[string]interface{} for custom configs from outside of library.

// Lexer is a set of settings that control lexeme scanning and comparison.
// And the methods for scanning and comparison are conviniently methods of
// Lexer.
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
// is either a string consisting of non-unicode.IsSpace characters,
// or a single newline character.
// If no lexemes found, nil is returned.
func (l *Lexer) Scan(text string) (xms []string) {
	r := strings.NewReader(text)
	s := bufio.NewScanner(r)
	s.Split(ScanLexemes)

	for s.Scan() {
		xms = append(xms, s.Text())
	}

	return
}

// Compare compares target against source and generates colored target's
// lexems highlighting mismatches between them. Additionally, actual
// comparison takes place between two non-LF lexems, and the spurious LFs
// are marked red and skipped. The function is intended to be called twice
// for the two permutations of the arguments to get error highlighting for
// both strings.
func (l *Lexer) Compare(target, source []string) (rts []RichText, ok bool) {
	rts = make([]RichText, len(target))
	ok = true

	ti, si := 0, 0
	for ; ti < len(target) && si < len(source); ti, si = ti+1, si+1 {
		// Skip spurious LFs
		if source[si] != "\n" {
			for ti < len(target) && target[ti] == "\n" {
				rts[ti].Str = "\n"
				rts[ti].Mask = []bool{true}
				ok = false
				ti++
			}
		} else if target[ti] != "\n" {
			for si < len(source) && source[si] == "\n" {
				si++
			}
		}

		if ti == len(target) || si == len(source) {
			break
		}

		xm := target[ti]
		rts[ti].Str = xm
		rts[ti].Mask = l.GenerateMask(xm, source[si])

		if rts[ti].Colorful() {
			ok = false
		}
	}

	for ; ti < len(target); ti++ {
		rts[ti].Str = target[ti]
		rts[ti].Mask = l.GenMaskForString(target[ti], "")

		ok = false
	}

	return
}

// deduceLexemeType will assess the type of the lexeme by sequentially applying
// more and more specialized type checkers starting from the least restrictive
// one.
func deduceLexemeType(xm string) lexemeType {
	for i := int(STRXM) + 1; i != int(FINALXM); i++ {
		// As any lexeme *is* a string, the function IsStringLexeme is omitted.
		if !TypeCheckers[i-1](xm) {
			return lexemeType(i - 1)
		}
	}

	return lexemeType(FINALXM - 1)
}

// GenerateMask is a wrapper function that finds the common type of the two
// lexems and generates a color mask for the target based on source.
func (l *Lexer) GenerateMask(target, source string) []bool {
	targetType := deduceLexemeType(target)
	sourceType := deduceLexemeType(source)

	commonType := targetType
	if sourceType < commonType {
		commonType = sourceType
	}

	return MaskGenerators[commonType](l, target, source)
}

// GenMaskForString will highlight mismatching characters.
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

// GenMaskForInt will highlight the whole number if at least one digit
// is different. Independently, the sign will be highlighted if it's different
// also.
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

// GenMaskForFloat uses the same logic as GenMaskForInt to highlight the
// whole part. If at least one digit in the fractional part (part after the
// dot) is different and its index (zero-based) is less than lexer's
// precision, this digit is highlighted.
func (l *Lexer) GenMaskForFloat(target, source string) (mask []bool) {
	targetWhole := strings.Split(target, ".")[0]

	sourceWhole := strings.Split(source, ".")[0]
	if sourceWhole == "" {
		sourceWhole = "0"
	}

	mask = l.GenMaskForInt(targetWhole, sourceWhole)

	if targetWhole == target {
		return
	}

	// dot is never colored
	mask = append(mask, false)

	// This one is never 0, because of the if up there that returns
	targetFracStart := strings.IndexRune(target, '.') + 1

	sourceFracStart := strings.IndexRune(source, '.') + 1
	if sourceFracStart == 0 {
		sourceFracStart = len(source)
	}

	targetFrac := target[targetFracStart:]
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
