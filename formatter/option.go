package formatter

type Opt func(c *Config)

func WithFormatterType(typ Type) Opt {
	return func(c *Config) {
		c.Type = typ
	}
}

func WithExternalFormatter(externalFormatter IFormatter) Opt {
	return func(c *Config) {
		c.ExternalFormatter = externalFormatter
	}
}

func WithTimestampFormat(format string) Opt {
	return func(c *Config) {
		c.TimestampFormat = format
	}
}

func WithPatternStyle(style string) Opt {
	return func(c *Config) {
		c.PatternStyle = style
	}
}

func WithEnableQuote(enable bool) Opt {
	return func(c *Config) {
		c.EnableQuote = enable
	}
}

func WithEnableQuoteEmptyFields(enable bool) Opt {
	return func(c *Config) {
		c.EnableQuoteEmptyFields = enable
	}
}

func WithDisableColors(disable bool) Opt {
	return func(c *Config) {
		c.DisableColors = disable
	}
}

func WithDisableTimestamp(disable bool) Opt {
	return func(c *Config) {
		c.DisableTimestamp = disable
	}
}

func WithDisableHTMLEscape(disable bool) Opt {
	return func(c *Config) {
		c.DisableHTMLEscape = disable
	}
}

func WithPrettyPrint(pretty bool) Opt {
	return func(c *Config) {
		c.PrettyPrint = pretty
	}
}
