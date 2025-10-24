package parser

// Option applies a modification to [ParserOptions] and returns it.
type Option func(p *ParserOptions) *ParserOptions

// WithStrictGrouping sets StrictGrouping on ParserOptions.
func WithStrictGrouping(val bool) Option {
	return func(opt *ParserOptions) *ParserOptions {
		opt.StrictGrouping = val
		return opt
	}
}

// WithAllowCurrencySymbol sets whether the parser accepts inputs
// that contain a currency symbol.
func WithAllowCurrencySymbol(val bool) Option {
	return func(opt *ParserOptions) *ParserOptions {
		opt.AllowCurrencySymbol = val
		return opt
	}
}

// WithAcceptSigns sets whether the parser accepts plus or minus signs in input.
func WithAcceptSigns(val bool) Option {
	return func(opt *ParserOptions) *ParserOptions {
		opt.AcceptSigns = val
		return opt
	}
}

// ParserOptions configures the Parser.
type ParserOptions struct {
	AllowCurrencySymbol bool
	StrictGrouping      bool
	AcceptSigns         bool
}

// DefaultOptions returns a [ParserOptions] with
//
// AllowCurrencySymbol=false, StrictGrouping=false, and AcceptSigns=true.
func DefaultOptions() *ParserOptions {
	return &ParserOptions{
		AllowCurrencySymbol: false,
		StrictGrouping:      false,
		AcceptSigns:         true,
	}
}
