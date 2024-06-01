package config

type Option func(cfg *Config)

func WithGeneralConfig(generalConfig *GeneralConfig) Option {
	return func(cfg *Config) {
		cfg.GeneralConfig = generalConfig
	}
}

func WithLogConfig(logConfig *LogConfig) Option {
	return func(cfg *Config) {
		cfg.LogConfig = logConfig
	}
}

func WithReportConfig(reportConfig *ReportConfig) Option {
	return func(cfg *Config) {
		cfg.ReportConfig = reportConfig
	}
}

type Config struct {
	*GeneralConfig
	*LogConfig
	*ReportConfig
}

func NewDefaultConfig() *Config {
	return &Config{
		GeneralConfig: NewGeneralConfig(),
		LogConfig:     NewLogConfig(),
		ReportConfig:  NewReportConfig(),
	}
}

func NewConfig(opts ...Option) *Config {
	cfg := NewDefaultConfig()
	for _, opt := range opts {
		opt(cfg)
	}
	
	return cfg
}
