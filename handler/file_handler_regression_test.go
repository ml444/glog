package handler

import "testing"

func newRegressionFileConfig(t *testing.T) *FileHandlerConfig {
	t.Helper()
	return &FileHandlerConfig{
		FileDir:       t.TempDir(),
		FileName:      "app",
		FileSuffix:    "log",
		MaxFileSize:   1 << 20,
		BufferSize:    8,
		BulkWriteSize: 256,
		RotatorType:   FileRotatorTypeSize,
		Backpressure:  BackpressureConfig{Strategy: BackpressureStrategyBlock},
	}
}
