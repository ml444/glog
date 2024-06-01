package formatter

type Option func(c *Config)

func WithFormatterType(typ Type) Option {
	return func(c *Config) {
		c.Type = typ
	}
}

func WithExternalFormatter(externalFormatter IFormatter) Option {
	return func(c *Config) {
		c.ExternalFormatter = externalFormatter
	}
}

func WithTimestampFormat(format string) Option {
	return func(c *Config) {
		c.TimestampFormat = format
	}
}

func WithPatternStyle(style string) Option {
	return func(c *Config) {
		c.PatternStyle = style
	}
}

func WithEnableQuote(enable bool) Option {
	return func(c *Config) {
		c.EnableQuote = enable
	}
}

func WithEnableQuoteEmptyFields(enable bool) Option {
	return func(c *Config) {
		c.EnableQuoteEmptyFields = enable
	}
}

func WithDisableColors(disable bool) Option {
	return func(c *Config) {
		c.DisableColors = disable
	}
}

func WithDisableTimestamp(disable bool) Option {
	return func(c *Config) {
		c.DisableTimestamp = disable
	}
}

func WithDisableHTMLEscape(disable bool) Option {
	return func(c *Config) {
		c.DisableHTMLEscape = disable
	}
}

func WithPrettyPrint(pretty bool) Option {
	return func(c *Config) {
		c.PrettyPrint = pretty
	}
}
