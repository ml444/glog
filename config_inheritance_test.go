package log

import (
	"strings"
	"testing"
)

func TestDefaultWorkerInheritsGlobalLevel(t *testing.T) {
	sink := &closeAwareHandler{}
	logger, err := New(SetLoggerLevel(DebugLevel), SetWorkerConfigs(NewDefaultStdoutWorkerConfig().SetHandler(sink)))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = logger.Stop() })
	logger.Debug("visible debug")
	if err := logger.Stop(); err != nil {
		t.Fatal(err)
	}
	if content, _ := sink.snapshot(); content != "visible debug\n" {
		t.Fatalf("default worker filtered Debug: %q", content)
	}
}

func TestExplicitWorkerLevelIsPreserved(t *testing.T) {
	sink := &closeAwareHandler{}
	worker := NewDefaultStdoutWorkerConfig().SetLevel(WarnLevel).SetHandler(sink)
	logger, err := New(SetLoggerLevel(DebugLevel), SetWorkerConfigs(worker))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = logger.Stop() })
	logger.Debug("filtered")
	logger.Error("accepted")
	if err := logger.Stop(); err != nil {
		t.Fatal(err)
	}
	if content, _ := sink.snapshot(); content != "accepted\n" {
		t.Fatalf("explicit worker level ignored: %q", content)
	}
}

func TestGlobalColorOverrideAndTimeInheritance(t *testing.T) {
	for _, color := range []bool{false, true} {
		for _, explicitLayout := range []string{"", "worker-time"} {
			stream := &closeAwareBuffer{}
			worker := NewDefaultStdoutWorkerConfig().SetStreamHandlerConfig(&StreamHandlerConfig{Streamer: stream})
			worker.FormatterCfg.Text.PatternStyle = "%[DateTime]s|%[LevelName]s|%[Message]v"
			worker.FormatterCfg.Text.TimeLayout = explicitLayout
			worker.FormatterCfg.Text.EnableColor = !color
			logger, err := New(SetWorkerConfigs(worker), SetTimeLayout("global-time"), SetColorRender(color))
			if err != nil {
				t.Fatal(err)
			}
			logger.Info("message")
			if err := logger.Stop(); err != nil {
				t.Fatal(err)
			}
			content, _ := stream.snapshot()
			wantLayout := explicitLayout
			if wantLayout == "" {
				wantLayout = "global-time"
			}
			if !strings.HasPrefix(content, wantLayout+"|") || strings.Contains(content, "\x1b[") != color {
				t.Fatalf("color=%t layout=%q: output=%q", color, explicitLayout, content)
			}
			if worker.FormatterCfg.Text.TimeLayout != explicitLayout || worker.FormatterCfg.Text.EnableColor != !color {
				t.Fatal("logger construction mutated the caller's formatter config")
			}
		}
	}
}

func TestImplicitFormatterUsesGlobalTimeLayout(t *testing.T) {
	stream := &closeAwareBuffer{}
	logger, err := New(SetTimeLayout("global-time"), SetWorkerConfigs(
		NewWorkerConfig(NoneLevel, 8).SetStreamHandlerConfig(&StreamHandlerConfig{Streamer: stream}),
	))
	if err != nil {
		t.Fatal(err)
	}
	logger.Info("message")
	if err := logger.Stop(); err != nil {
		t.Fatal(err)
	}
	if content, _ := stream.snapshot(); !strings.Contains(content, "global-time") {
		t.Fatalf("implicit formatter ignored global layout: %q", content)
	}
}

func TestNewLoggerDoesNotNormalizeCallerConfig(t *testing.T) {
	stream := &closeAwareBuffer{}
	worker := NewWorkerConfig(NoneLevel, 0).SetStreamHandlerConfig(&StreamHandlerConfig{Streamer: stream})
	cfg := &Config{WorkerConfigList: []*WorkerConfig{worker}}
	logger, err := NewLogger(cfg)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = logger.Stop() })
	if cfg.LoggerLevel != NoneLevel || cfg.TimeLayout != "" || cfg.OnError != nil || worker.CacheSize != 0 || worker.FormatterCfg.Text != nil {
		t.Fatal("NewLogger modified the supplied config")
	}
}

func TestInitLogReappliesInheritedSettings(t *testing.T) {
	isolateGlobalLogger(t)
	stream := &closeAwareBuffer{}
	worker := NewDefaultStdoutWorkerConfig().SetStreamHandlerConfig(&StreamHandlerConfig{Streamer: stream})
	worker.FormatterCfg.Text.PatternStyle = "%[DateTime]s|%[LoggerName]s|%[Message]v"
	if err := InitLog(SetWorkerConfigs(worker), SetColorRender(false), SetLoggerName("first"), SetTimeLayout("first-time")); err != nil {
		t.Fatal(err)
	}
	Info("one")
	if err := InitLog(SetLoggerName("next"), SetTimeLayout("next-time")); err != nil {
		t.Fatal(err)
	}
	Info("two")
	Stop()
	if content, _ := stream.snapshot(); content != "first-time|first|one\nnext-time|next|two\n" {
		t.Fatalf("inherited settings froze during reinitialization: %q", content)
	}
}

func TestFailedInitLogPreservesConfigAndAllowsRetry(t *testing.T) {
	isolateGlobalLogger(t)
	sink := &closeAwareHandler{}
	if err := InitLog(SetWorkerConfigs(NewWorkerConfig(NoneLevel, 8).SetHandler(sink))); err != nil {
		t.Fatal(err)
	}
	originalConfig, originalLogger := Conf, GetLogger()
	if err := InitLog(SetWorkerConfigs(NewWorkerConfig(NoneLevel, -1).SetHandler(sink))); err == nil {
		t.Fatal("negative cache size unexpectedly accepted")
	}
	if Conf != originalConfig || GetLogger() != originalLogger {
		t.Fatal("failed initialization changed active state")
	}
	if err := InitLog(); err != nil {
		t.Fatalf("retry inherited an invalid config: %v", err)
	}
	Info("after retry")
	Stop()
	if content, _ := sink.snapshot(); content != "after retry\n" {
		t.Fatalf("retry failed to preserve output: %q", content)
	}
}
