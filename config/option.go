package config

type OptionFunc func (config *Config) error


func SetLoggerName(name string) OptionFunc {
	return func(config *Config) error {
		config.Logger.Name = name
		return nil
	}
}

func SetFileName(name string) OptionFunc {
	return func(config *Config) error {
		config.Handler.LogHandlerConfig.File.FileName = name
		return nil
	}
}

func SetFileDir(path string) OptionFunc {
	return func(config *Config) error {
		config.Handler.LogHandlerConfig.File.FileDir = path
		return nil
	}
}
