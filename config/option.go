package config

type Option func(cfg *Config)

var WithGeneralOption = func(cfg *Config, opt GeneralOption) {
	opt(cfg.GeneralConfig)
}

var WithLogOption = func(cfg *Config, opt LogOption) {
	opt(cfg.LogConfig)
}

var WithReportOption = func(cfg *Config, opt ReportOption) {
	opt(cfg.ReportConfig)
}

type GeneralOption func(cfg *GeneralConfig)

type LogOption func(cfg *LogConfig)

type ReportOption func(cfg *ReportConfig)
