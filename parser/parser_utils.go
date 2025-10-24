package parser

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/Rhymond/go-money"
)

func lookupCurrency(q string) (*money.Currency, error) {
	if isAlpha3(q) {
		if c := money.GetCurrency(q); c != nil {
			return c, nil
		}
		return nil, fmt.Errorf("%w: %s", ErrInvalidISO, q)
	}

	if isNumeric(q) {
		if c := money.GetCurrencyByNumericCode(q); c != nil {
			return c, nil
		}
		return nil, fmt.Errorf("%w: %s", ErrInvalidNumericCode, q)
	}

	return nil, fmt.Errorf("%w: %q", ErrInvalidCurrencyQuery, q)
}

func isAlpha3(s string) bool {
	if len(s) != 3 {
		return false
	}
	for i := 0; i < 3; i++ {
		ch := s[i]
		if ('a' > ch || ch > 'z') && ('A' > ch || ch > 'Z') {
			return false
		}
	}
	return true
}

func isNumeric(s string) bool {
	if len(s) == 0 {
		return false
	}
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return true
}

func containsCurrencySymbol(s string) bool {
	for i := 0; i < len(s); {
		r, sz := utf8.DecodeRuneInString(s[i:])
		if unicode.Is(unicode.Sc, r) {
			return true
		}
		i += sz
	}

	lc := strings.ToLower(s)

	tokens := [...]string{
		".د.إ", ".د.ب", ".د.ت", ".د.ج", ".د.ع", ".د.ك", ".د.ل", ".د.م",
		"a$", "ar", "b/.", "br", "bs", "bs.", "bs.s", "bz$", "c$", "cf",
		"cfa", "cg", "chf", "d", "db", "fc", "fdj", "fg", "fr", "frw",
		"ft", "g", "gs", "hk$", "j$", "k", "km", "kn", "kr", "ksh", "kz",
		"kč", "l", "le", "lei", "ls", "lt", "mk", "mt", "mvr", "nfk", "nt$",
		"nu.", "oz t", "p", "p.", "q", "r", "r$", "rd$", "rm", "rp", "s$",
		"s/", "sdr", "sh", "sk", "sm", "so’m", "t", "t$", "tsh", "tt$", "uf",
		"um", "ush", "vt", "z$", "zk", "zł", "ƒ", "ден", "дин.", "лв", "сом",
		"դր.", "ლ", "元",
	}

	for _, t := range tokens {
		if strings.Contains(lc, t) {
			return true
		}
	}

	return false
}

func containsSign(s string) bool {
	allowed := []rune{'-', '+', '−'}
	r, _ := utf8.DecodeRuneInString(s)
	return contains(allowed, r)
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

func pow10int64(n int) int64 {
	switch n {
	case 0:
		return 1
	case 1:
		return 10
	case 2:
		return 100
	case 3:
		return 1000
	case 4:
		return 10000
	case 5:
		return 100000
	case 6:
		return 1000000
	case 7:
		return 10000000
	case 8:
		return 100000000
	case 9:
		return 1000000000
	default:
		panic("fraction digits out of range")
	}
}

func contains(slice []rune, r rune) bool {
	for _, c := range slice {
		if c == r {
			return true
		}
	}

	return false
}
