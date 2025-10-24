package parser

import (
	"errors"
	"testing"

	"github.com/Rhymond/go-money"
)

type tc struct {
	name string
	in   string
	iso  string
	want int64
	opts []Option
	err  error
}

func TestParseAmount_Table(t *testing.T) {
	t.Parallel()

	cases := []tc{
		{name: "ok/JPY/100", in: "100", iso: money.JPY, want: 100},
		{name: "ok/BGN/100", in: "100", iso: money.BGN, want: 10000},
		{name: "ok/BGN/100.0", in: "100.0", iso: money.BGN, want: 10000},
		{name: "ok/BGN/100.00", in: "100.00", iso: money.BGN, want: 10000},
		{name: "ok/EUR/100.00", in: "100.00", iso: money.EUR, want: 10000},

		{name: "ok/EUR/100,000.00", in: "100,000.00", iso: money.EUR, want: 10000000},
		{name: "ok/EUR/100 000.00", in: "100 000.00", iso: money.EUR, want: 10000000},
		{name: "ok/EUR/NBSP-100000.00", in: "100\u00A0000.00", iso: money.EUR, want: 10000000},

		{name: "ok/CLF/2,28", in: "2,28", iso: money.CLF, want: 22800},

		{name: "ok/LYD/539.72-pad", in: "539.72", iso: money.LYD, want: 539720},
		{name: "ok/LYD/539-no-dec", in: "539", iso: money.LYD, want: 539000},

		{name: "ok/LYD/+539/signs-allowed", in: "+539", iso: money.LYD, want: 539000, opts: []Option{WithAcceptSigns(true)}},

		{name: "ok/USD/$539/symbol-allowed", in: "$539", iso: money.USD, want: 53900, opts: []Option{WithAllowCurrencySymbol(true)}},

		{name: "err/USD/10 000,000.00/strict-grouping-enabled", in: "10 000,000.00", iso: money.USD, opts: []Option{WithStrictGrouping(true)}, err: ErrMixedGrouping},
		{name: "err/USD/10 000,000.00/strict-grouping-disabled", in: "10 000,000.00", iso: money.USD, want: 1000000000, opts: []Option{WithStrictGrouping(false)}},

		{name: "err/empty/empty-input", in: "", iso: money.EUR, err: ErrEmptyInput},
		{name: "err/iso/empty-iso", in: "1", iso: "", err: ErrInvalidISO},
		{name: "err/iso/unknown-iso", in: "1", iso: "ZZZ", err: ErrInvalidISO},

		{name: "err/LYD/+539/signs-not-allowed", in: "+539", iso: money.LYD, opts: []Option{WithAcceptSigns(false)}, err: ErrSignsNotAllowed},
		{name: "err/LYD/-539/signs-not-allowed", in: "-539", iso: money.LYD, opts: []Option{WithAcceptSigns(false)}, err: ErrSignsNotAllowed},

		// symbol handling
		{name: "err/USD/$539/symbols-not-allowed", in: "$539", iso: money.USD, opts: []Option{WithAllowCurrencySymbol(false)}, err: ErrCurrencySymbolNotAllowed},
		{name: "err/USD/€539/symbol-mismatch", in: "\u20ac539", iso: money.USD, opts: []Option{WithAllowCurrencySymbol(true)}, err: ErrInvalidCurrencySymbol},

		{name: "err/EUR/1.234/too-many-fraction", in: "1.234", iso: money.EUR, err: ErrTooManyDecimals},

		{name: "err/EUR/12a3/bad-char", in: "12a3", iso: money.EUR, err: ErrBadChar},

		{name: "err/EUR/only-plus/no-digits", in: "+", iso: money.EUR, opts: []Option{WithAcceptSigns(true)}, err: ErrNoDigits},
		{name: "err/USD/only-symbol/no-digits", in: "$", iso: money.USD, opts: []Option{WithAllowCurrencySymbol(true)}, err: ErrNoDigits},

		{name: "ok/USD/supports-pkg-formatted/USD/1,234,567.89", in: "1,234,567.89 $", iso: money.USD, want: 123456789, opts: []Option{WithAllowCurrencySymbol(true), WithStrictGrouping(true)}},
		{name: "ok/USD/supports-pkg-formatted/GBP/1,234,567.89", in: "£1,234,567.89", iso: money.GBP, want: 123456789, opts: []Option{WithAllowCurrencySymbol(true), WithStrictGrouping(true)}},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			c := c
			t.Parallel()

			p := NewAmountParser(c.opts...)
			got, err := p.Parse(c.in, c.iso)

			if c.err != nil {
				if err == nil || !errors.Is(err, c.err) {
					t.Fatalf("ParseAmount(%q,%q) error = %v, want errors.Is(...,%v)", c.in, c.iso, err, c.err)
				}
				return
			}

			if err != nil {
				t.Fatalf("ParseAmount(%q,%q) unexpected error: %v", c.in, c.iso, err)
			}
			if got != c.want {
				t.Errorf("ParseAmount(%q,%q) = %d, want %d", c.in, c.iso, got, c.want)
			}
		})
	}
}

func TestParseAmount_StripsNBSPAndSpaces(t *testing.T) {
	t.Parallel()

	p := NewAmountParser()
	got, err := p.Parse(" \t\u00a0100\u00a0.00 ", money.EUR)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != 10000 {
		t.Fatalf("got %d, want %d", got, 10000)
	}
}

func TestParseAmount_RejectsAnyCurrencySymbolWhenNotAllowed(t *testing.T) {
	t.Parallel()

	p := NewAmountParser(WithAllowCurrencySymbol(false))
	_, err := p.Parse("100¤", money.EUR)
	if !errors.Is(err, ErrCurrencySymbolNotAllowed) {
		t.Fatalf("err = %v, want ErrCurrencySymbolNotAllowed", err)
	}
}
