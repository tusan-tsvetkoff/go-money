// Package parser provides a way to turn strings into [money.Amount], given a valid Alpha-3 ISO
// or numeric currency code.
//
// ref: https://en.wikipedia.org/wiki/List_of_ISO_3166_country_codes
//
// The [Parser] interface allows for custom parsing implementations and easier testing.
//
// Calling [NewAmountParser] returns the [Parser] with default [ParserOptions].
// An example of converting the string value of "1,455.00" into a money.Amount of 145500:
//
//	parser := parser.NewMoneyParser()
//	amount, err := parser.ParseAmount("1,455.00", "EUR") // 145500
//
// By default, the parser accepts signs (+/-):
//
//	parser := parser.NewMoneyParser()
//	negativeAmount, err := parser.ParseAmount("-1,455.00", "EUR") // -145500
//	positiveAmount, err := parser.ParseAmount("+1,455.00", "EUR") // 145500
//
// to disable them, you must use an option:
//
//	p := parser.NewMoneyParser(WithAcceptSigns(false))
//	_, err := p.ParseAmount("-1,455.00", "EUR") // [ErrSignsNotAllowed]
//
// Using [ParserOptions], you can also enable the use of currency symbols:
//
//	parser := parser.NewMoneyParser(WithAllowCurrencySymbol(true))
//	amount, err := parser.ParseAmount("€1,455.00", "EUR") // 145500
//
// Note: If a currency symbol is used in the input string, it must match the
// currency corresponding to the provided ISO code; otherwise, an error is returned.
package parser

import (
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/Rhymond/go-money"
)

const (
	nbsp  = "\u00A0"
	space = " "

	minusSign  = '−'
	plusSign   = '+'
	hyphenSign = '-'
)

var (
	// ErrEmptyInput is returned when the input string is empty.
	ErrEmptyInput = errors.New("empty input")
	// ErrSignsNotAllowed is returned when the input string has a sign but signs are not allowed.
	ErrSignsNotAllowed = errors.New("signs not allowed")
	// ErrCurrencySymbolNotAllowed is returned when the input string has a currency symbol but currency symbols are not allowed.
	ErrCurrencySymbolNotAllowed = errors.New("currency symbol not allowed")
	// ErrMixedGrouping is returned when the input string has mixed thousands grouping characters.
	ErrMixedGrouping = errors.New("mixed grouping")
	// ErrInvalidCurrencySymbol is returned when the currency symbol in the input string does not match the given ISO code.
	ErrInvalidCurrencySymbol = errors.New("invalid currency symbol")
	// ErrInvalidISO is returned when the given ISO code is invalid.
	ErrInvalidISO = errors.New("invalid ISO currency")
	// ErrInvalidNumericCode is returned when the given numeric code invalid.
	ErrInvalidNumericCode = errors.New("invalid numeric code")
	// ErrInvalidCurrencyQuery is returned when the query is neither an ISO or a numeric code.
	ErrInvalidCurrencyQuery = errors.New("invalid currency querry")
	// ErrTooManyDecimals is returned when the input string has too many fractional digits for the given ISO code.
	ErrTooManyDecimals = errors.New("too many fractional digits")
	// ErrBadChar is returned when the input string contains an invalid character.
	ErrBadChar = errors.New("invalid character")
	// ErrNoDigits is returned when the input string contains no digits.
	ErrNoDigits = errors.New("no digits")
)

// Parser is the interface for parsing monetary strings.
type Parser interface {
	// Parse parses a string into a [money.Amount] based on the given
	// ISO or numeric code and input.
	Parse(s, currency string) (money.Amount, error)
}

// AmountParser is the default implementation of the [Parser] interface.
type AmountParser struct {
	opt ParserOptions
}

var _ Parser = (*AmountParser)(nil)

// NewAmountParser returns a new [AmountParser] with the given options.
func NewAmountParser(opts ...Option) *AmountParser {
	mp := &AmountParser{}
	opt := &ParserOptions{}
	if len(opts) > 0 {
		for _, o := range opts {
			opt = o(opt)
		}
	} else {
		opt = DefaultOptions()
	}

	mp.opt = *opt

	return mp
}

// Parse parses a string into a [money.Amount] based on the given
// ISO code and input.
func (p *AmountParser) Parse(input string, currency string) (money.Amount, error) {
	s := strings.TrimSpace(input)
	if s == "" {
		return money.AmountZero, ErrEmptyInput
	}
	if currency == "" {
		return money.AmountZero, ErrInvalidISO
	}

	q := strings.TrimSpace(currency)
	c, err := lookupCurrency(q)
	if err != nil {
		return money.AmountZero, err
	}

	if !p.opt.AcceptSigns && containsSign(s) {
		return money.AmountZero, fmt.Errorf("input %q: %w", s, ErrSignsNotAllowed)
	}
	if !p.opt.AllowCurrencySymbol && containsCurrencySymbol(s) {
		return money.AmountZero, fmt.Errorf("input %q: %w", s, ErrCurrencySymbolNotAllowed)
	}

	return p.parse(s, *c)
}

func (p *AmountParser) parse(s string, cur money.Currency) (money.Amount, error) {
	s = strings.TrimSpace(strings.ReplaceAll(s, nbsp, space))

	if p.opt.AllowCurrencySymbol && len(s) > 0 {
		currIdx := strings.Index(s, cur.Grapheme)
		if currIdx == -1 {
			return money.AmountZero, ErrInvalidCurrencySymbol
		}

		s = strings.Replace(s, cur.Grapheme, "", 1)
		s = strings.TrimSpace(s)
	}

	var sign int64 = 1
	if p.opt.AcceptSigns && len(s) > 0 {
		r, size := utf8.DecodeRuneInString(s)
		switch r {
		case minusSign, hyphenSign:
			sign = -1
			s = strings.TrimSpace(s[size:])
		case plusSign:
			s = strings.TrimSpace(s[size:])
		}
	}

	if s == "" {
		return money.AmountZero, ErrNoDigits
	}

	dec := rune('.')
	if len(cur.Decimal) > 0 {
		dec = []rune(cur.Decimal)[0]
	}

	fracDigits := cur.Fraction

	var intDigits, fracDigitsRunes []rune
	hasDec := false

	lastSeenRune := rune(48)
	for _, r := range s {
		switch {
		case r >= '0' && r <= '9':
			if hasDec {
				fracDigitsRunes = append(fracDigitsRunes, r)
			} else {
				intDigits = append(intDigits, r)
			}
		case r == dec && !hasDec && fracDigits > 0:
			hasDec = true
		case r == ' ' || r == ',' || r == '.':
			if p.opt.StrictGrouping {
				tmp := lastSeenRune
				lastSeenRune = r
				if !hasDec && (tmp != 48 && tmp != lastSeenRune) {
					return money.AmountZero, fmt.Errorf("input: %s: %w: %c", s, ErrMixedGrouping, r)
				}
			}
			continue
		default:
			return 0, fmt.Errorf("%w: %q", ErrBadChar, r)
		}
	}

	if len(intDigits) == 0 && len(fracDigitsRunes) == 0 {
		return 0, ErrNoDigits
	}

	switch {
	case len(fracDigitsRunes) < fracDigits:
		for i := len(fracDigitsRunes); i < fracDigits; i++ {
			fracDigitsRunes = append(fracDigitsRunes, '0')
		}
	case len(fracDigitsRunes) > fracDigits:
		return 0, ErrTooManyDecimals
	}

	intVal, err := atoiRunes(intDigits)
	if err != nil {
		return 0, err
	}
	fracVal, err := atoiRunes(fracDigitsRunes)
	if err != nil {
		return 0, err
	}

	base := pow10int64(fracDigits)
	minor := intVal*base + fracVal

	return money.Amount(sign * minor), nil
}
