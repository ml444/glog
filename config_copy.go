package log

import "fmt"

func (c *Config) clone() *Config {
	if c == nil {
		return NewDefaultConfig()
	}
	copyConfig := *c
	if c.EnableColorRender != nil {
		value := *c.EnableColorRender
		copyConfig.EnableColorRender = &value
	}
	if c.EnableRoutineID != nil {
		value := *c.EnableRoutineID
		copyConfig.EnableRoutineID = &value
	}
	if c.WorkerConfigList != nil {
		copyConfig.WorkerConfigList = make([]*WorkerConfig, len(c.WorkerConfigList))
		for index, worker := range c.WorkerConfigList {
			if worker == nil {
				continue
			}
			copyWorker := *worker
			if worker.HandlerCfg.File != nil {
				value := *worker.HandlerCfg.File
				copyWorker.HandlerCfg.File = &value
			}
			if worker.HandlerCfg.Stream != nil {
				value := *worker.HandlerCfg.Stream
				copyWorker.HandlerCfg.Stream = &value
			}
			if worker.HandlerCfg.Syslog != nil {
				value := *worker.HandlerCfg.Syslog
				copyWorker.HandlerCfg.Syslog = &value
			}
			if worker.FormatterCfg.Text != nil {
				value := *worker.FormatterCfg.Text
				copyWorker.FormatterCfg.Text = &value
			}
			if worker.FormatterCfg.JSON != nil {
				value := *worker.FormatterCfg.JSON
				copyWorker.FormatterCfg.JSON = &value
			}
			if worker.FormatterCfg.XML != nil {
				value := *worker.FormatterCfg.XML
				copyWorker.FormatterCfg.XML = &value
			}
			copyConfig.WorkerConfigList[index] = &copyWorker
		}
	}
	return &copyConfig
}

func (c *Config) validate() error {
	for index, worker := range c.WorkerConfigList {
		if worker.CacheSize < 0 {
			return fmt.Errorf("worker %d cache size must not be negative", index)
		}
		if file := worker.HandlerCfg.File; worker.CustomHandler == nil && file != nil && file.BufferSize < 0 {
			return fmt.Errorf("worker %d file buffer size must not be negative", index)
		}
	}
	return nil
}
