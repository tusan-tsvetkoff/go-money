package parser

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/Rhymond/go-money"
)

const (
	nbsp  = "\u00A0"
	space = " "

	nbspR  = '\u00A0'
	spaceR = ' '
)

var (
	errEmptyInput          = errors.New("cannot parse an empty string")
	errISOCodeNotSpecified = errors.New("ISO code must be specified")
	errInvalidISOCode      = errors.New("invalid ISO code")
	errNoDigits            = errors.New("no digits")
)

func ParseToMinor(input, isoCode string) (money.Amount, error) {
	s := strings.TrimSpace(input)
	if s == "" {
		return 0, errEmptyInput
	}

	if isoCode == "" {
		return 0, errISOCodeNotSpecified
	}

	c := money.GetCurrency(isoCode)
	if c == nil {
		return 0, fmt.Errorf("%v: %s", errInvalidISOCode, isoCode)
	}

	return parseToMinor(s, *c)
}

func parseToMinor(s string, c money.Currency) (money.Amount, error) {
	sign := int64(1)
	switch s[0] {
	case '+':
		s = strings.TrimSpace(s[1:])
	case '-':
		sign = -1
		s = strings.TrimSpace(s[1:])
	}
	if s == "" {
		return money.AmountZero, errNoDigits
	}

	s = strings.ReplaceAll(s, nbsp, space)

	decimable := c.Fraction > 0
	if decimable {
		dec := c.Decimal
		fraction := c.Fraction

		intPart := s
		fracPart := ""

		decIdx := strings.LastIndexByte(s, byte(dec[0]))
		if decIdx != -1 {
			fracPart = s[decIdx+1:]
			intPart = s[:decIdx]
		}

		fracLenPrePad := runeCount(fracPart)
		if fracLenPrePad < c.Fraction {
			diff := c.Fraction - fracLenPrePad
			fracPart = fracPart + strings.Repeat("0", diff)
		}

		intSequence, err := normalizeSequence(intPart, IntSequence)
		if err != nil {
			return money.AmountZero, err
		}

		fracSequence, err := normalizeSequence(fracPart, DecSequence)
		if err != nil {
			return money.AmountZero, err
		}

		intVal, err := atoiRunes(intSequence)
		if err != nil {
			return money.AmountZero, err
		}

		fracVal, err := atoiRunes(fracSequence)
		if err != nil {
			return money.AmountZero, err
		}

		var base int64 = 1
		for range fraction {
			base *= 10
		}

		minor := intVal*base + fracVal

		amount := money.Amount(sign * minor)

		return amount, nil
	}

	intPart := s
	fraction := c.Fraction

	cl, err := normalizeSequence(intPart, IntSequence)
	if err != nil {
		return money.AmountZero, err
	}

	intVal, err := atoiRunes(cl)
	if err != nil {
		return 0, err
	}

	var base int64 = 1
	for range fraction {
		base *= 10
	}

	minor := intVal * base

	amount := money.Amount(sign * minor)

	return amount, nil
}

type sequence int

const (
	IntSequence sequence = iota
	DecSequence
)

func normalizeSequence(seq string, t sequence) ([]rune, error) {
	var length int
	switch t {
	case IntSequence:
		seq = strings.ReplaceAll(seq, ",", "")
		seq = strings.ReplaceAll(seq, ".", "")
		seq = strings.ReplaceAll(seq, " ", "")

		var convErr error
		length, convErr = strconv.Atoi(seq)
		if convErr != nil {
			return nil, convErr
		}
	case DecSequence:
		length = runeCount(seq)
	default:
		return nil, fmt.Errorf("invalid sequence=%#v", t)
	}

	return normalizeSequenceInternal(length, seq)
}

func normalizeSequenceInternal(length int, sequence string) ([]rune, error) {
	cleanInt := make([]rune, 0, length)

	for _, r := range sequence {
		if unicode.IsDigit(r) {
			cleanInt = append(cleanInt, r)
			continue
		}

		switch r {
		case spaceR, nbspR:
			continue
		default:
			return nil, fmt.Errorf("invalid character in sequence: %c", r)
		}
	}

	return cleanInt, nil
}

func atoiRunes(rs []rune) (int64, error) {
	var n int64
	for _, r := range rs {
		if r < '0' || r > '9' {
			return 0, fmt.Errorf("not a digit: %q", r)
		}
		n = n*10 + int64(r-'0')
	}
	return n, nil
}

func runeCount(s string) int {
	return utf8.RuneCountInString(s)
}
