package parser_test

import (
	"testing"

	"github.com/Rhymond/go-money/parser"
)

func FuzzParseStringToAmount(f *testing.F) {
	f.Add("1,234.56", "USD", false, false, false)
	f.Add("1 234,56", "EUR", false, false, false)
	f.Add("-100", "JPY", false, true, false)
	f.Add("  +0  ", "USD", false, true, false)
	f.Add("not-a-number", "USD", false, false, false)
	f.Add("-500", "JPY", false, true, false)
	f.Add("1  ", "BGN", false, false, false)
	f.Add("2 450 000.34", "AUD", false, false, false)
	f.Add("$1,234.55", "USD", true, false, false)
	f.Add("€1,234.55", "EUR", true, false, false)
	f.Add("€1,234.55", "BGN", true, false, false)
	f.Add("лв1,234.55", "975", true, false, false)
	f.Add("\u043b\u04321,234.55", "BGN", true, false, false)
	f.Add("€1,234.55", "BGN", false, false, false)
	f.Add("1 000,234.55", "BGN", false, false, true)

	f.Fuzz(func(t *testing.T, s, iso string, currencySymbol, acceptSigns, strictGrouping bool) {
		var opts []parser.Option
		if currencySymbol {
			opts = append(opts, parser.WithAllowCurrencySymbol(true))
		}
		if acceptSigns {
			opts = append(opts, parser.WithAcceptSigns(true))
		}
		if strictGrouping {
			opts = append(opts, parser.WithStrictGrouping(true))
		}

		p := parser.NewAmountParser(opts...)
		_, _ = p.Parse(s, iso)
	})
}
