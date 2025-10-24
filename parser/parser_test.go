package parser

import (
	"testing"

	"github.com/Rhymond/go-money"
)

func Test_ParseToMinor(t *testing.T) {
	cases := []struct {
		given string
		iso   string
		want  int64
		desc  string
	}{
		{
			given: "100",
			iso:   money.JPY,
			want:  100,
			desc:  "Handle no decimal currencies",
		},
		{
			given: "100",
			iso:   money.BGN,
			want:  10000,
			desc:  "Handle BGN no decimal",
		},
		{
			given: "100.0",
			iso:   money.BGN,
			want:  10000,
			desc:  "Handle BGN one decimal less",
		},
		{
			given: "100.00",
			iso:   money.BGN,
			want:  10000,
			desc:  "Handle BGN",
		},
		{
			given: "100.00",
			iso:   money.EUR,
			want:  10000,
			desc:  "Handle EUR",
		},
		{
			given: "100,000.00",
			iso:   money.EUR,
			want:  10000000,
			desc:  "Handle EUR thousands",
		},
		{
			given: "100 000.00",
			iso:   money.EUR,
			want:  10000000,
			desc:  "Handle EUR thousands with space",
		},
		{
			given: "2,28",
			iso:   money.CLF,
			want:  22800,
			desc:  "Handle CLF",
		},
		{
			given: "539.72",
			iso:   money.LYD,
			want:  539720,
			desc:  "Handles LYD",
		},
		{
			given: "539",
			iso:   money.LYD,
			want:  539000,
			desc:  "Handles LYD no dec given",
		},
	}
	for _, c := range cases {
		c := c // capture
		t.Run(c.desc, func(t *testing.T) {
			got, err := ParseToMinor(c.given, c.iso)
			if err != nil {
				t.Fatalf("ParseToMinor(%q, %q) unexpected error: %v", c.given, c.iso, err)
			}
			if got != c.want {
				t.Errorf("ParseToMinor(%q, %q): got %d; want %d", c.given, c.iso, got, c.want)
			}
		})
	}
}
