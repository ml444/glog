package log

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/ml444/glog/formatter"
	"github.com/ml444/glog/message"
)

type failingEmitHandler struct{}

func (failingEmitHandler) Emit(*message.Entry) error { return errors.New("emit failed") }
func (failingEmitHandler) Close() error              { return nil }

func TestEmitErrorCallbackCanStopLogger(t *testing.T) {
	callbackDone := make(chan error, 1)
	var logger *Logger
	var err error
	logger, err = New(SetWorkerConfigs(NewWorkerConfig(NoneLevel, 8).SetHandler(failingEmitHandler{})),
		SetOnError(func(interface{}, error) { callbackDone <- logger.Stop() }))
	if err != nil {
		t.Fatal(err)
	}
	logger.Info("trigger")
	select {
	case err := <-callbackDone:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(time.Second):
		t.Fatal("emit error callback deadlocked stopping its logger")
	}
}

type failingRecordFormatter struct{}

func (failingRecordFormatter) Format(*message.Entry) ([]byte, error) {
	return nil, errors.New("format failed")
}

func TestFileErrorCallbackCanStopLogger(t *testing.T) {
	callbackDone := make(chan error, 1)
	var logger *Logger
	var err error
	worker := NewDefaultTextFileWorkerConfig(t.TempDir()).SetFormatter(failingRecordFormatter{})
	logger, err = New(SetWorkerConfigs(worker),
		SetOnError(func(interface{}, error) { callbackDone <- logger.Stop() }))
	if err != nil {
		t.Fatal(err)
	}
	logger.Info("trigger")
	select {
	case err := <-callbackDone:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(time.Second):
		t.Fatal("file error callback deadlocked stopping its logger")
	}
}

func TestStopDoesNotWaitForOutputErrorCallbacks(t *testing.T) {
	started := make(chan struct{})
	release := make(chan struct{})
	defer close(release)
	logger, err := New(SetWorkerConfigs(NewWorkerConfig(NoneLevel, 8).SetHandler(failingEmitHandler{})),
		SetOnError(func(interface{}, error) {
			close(started)
			<-release
		}))
	if err != nil {
		t.Fatal(err)
	}
	logger.Info("trigger")
	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("callback did not start")
	}
	stopped := make(chan error, 1)
	go func() { stopped <- logger.Stop() }()
	select {
	case err := <-stopped:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(time.Second):
		t.Fatal("Stop waited for an output error callback")
	}
}

func TestWorkerSettersReplaceMutuallyExclusiveChoices(t *testing.T) {
	worker := NewDefaultTextFileWorkerConfig(t.TempDir())
	streamConfig := &StreamHandlerConfig{Streamer: &closeAwareBuffer{}}
	worker.SetStreamHandlerConfig(streamConfig)
	if worker.HandlerCfg.File != nil || worker.HandlerCfg.Syslog != nil || worker.HandlerCfg.Stream != streamConfig {
		t.Fatal("stream setter did not replace file config")
	}
	worker.SetHandler(failingEmitHandler{})
	if worker.HandlerCfg != (HandlerConfig{}) || worker.CustomHandler == nil {
		t.Fatal("custom handler did not replace built-in config")
	}
	worker.SetSyslogHandlerConfig(&SyslogHandlerConfig{})
	if worker.CustomHandler != nil || worker.HandlerCfg.Syslog == nil || worker.HandlerCfg.Stream != nil {
		t.Fatal("syslog setter did not replace custom handler")
	}
	worker.SetFileHandlerConfig(NewDefaultFileHandlerConfig(t.TempDir()))
	if worker.HandlerCfg.File == nil || worker.HandlerCfg.Syslog != nil {
		t.Fatal("file setter did not replace syslog config")
	}
	worker.SetJSONFormatterConfig(NewDefaultJSONFormatterConfig())
	if worker.FormatterCfg.Text != nil || worker.FormatterCfg.XML != nil || worker.FormatterCfg.JSON == nil {
		t.Fatal("JSON setter did not replace text formatter")
	}
	worker.SetXMLFormatterConfig(&XMLFormatterConfig{})
	if worker.FormatterCfg.JSON != nil || worker.FormatterCfg.XML == nil {
		t.Fatal("XML setter did not replace JSON formatter")
	}
	worker.SetFormatter(formatter.NewTextFormatter(formatter.TextFormatterConfig{}))
	if worker.FormatterCfg != (FormatterConfig{}) || worker.CustomFormatter == nil {
		t.Fatal("custom formatter did not replace built-in config")
	}
	worker.SetTextFormatterConfig(NewDefaultTextFormatterConfig())
	if worker.CustomFormatter != nil || worker.FormatterCfg.Text == nil {
		t.Fatal("text setter did not replace custom formatter")
	}
}

func TestDefaultFileWorkerCanSwitchToJSONStream(t *testing.T) {
	stream := &closeAwareBuffer{}
	worker := NewDefaultTextFileWorkerConfig(t.TempDir()).
		SetStreamHandlerConfig(&StreamHandlerConfig{Streamer: stream}).
		SetJSONFormatterConfig(NewDefaultJSONFormatterConfig())
	logger, err := New(SetWorkerConfigs(worker), SetColorRender(false))
	if err != nil {
		t.Fatal(err)
	}
	logger.Info("payload")
	if err := logger.Stop(); err != nil {
		t.Fatal(err)
	}
	content, _ := stream.snapshot()
	var record message.Record
	if err := json.Unmarshal([]byte(content), &record); err != nil || record.Message != "payload" {
		t.Fatalf("stream output = %q, decode error = %v", content, err)
	}
}

func TestCallerLocationForInstanceAndPackageAPIs(t *testing.T) {
	isolateGlobalLogger(t)
	methods := []struct {
		name string
		emit func(*Logger) (string, int)
	}{
		{"instance Info", func(logger *Logger) (string, int) {
			_, file, line, _ := runtime.Caller(0)
			logger.Info("payload")
			return file, line + 1
		}},
		{"instance Infof", func(logger *Logger) (string, int) {
			_, file, line, _ := runtime.Caller(0)
			logger.Infof("%s", "payload")
			return file, line + 1
		}},
		{"instance Log", func(logger *Logger) (string, int) {
			_, file, line, _ := runtime.Caller(0)
			logger.Log(InfoLevel, "payload")
			return file, line + 1
		}},
		{"instance Logf", func(logger *Logger) (string, int) {
			_, file, line, _ := runtime.Caller(0)
			logger.Logf(InfoLevel, "%s", "payload")
			return file, line + 1
		}},
		{"package Info", func(*Logger) (string, int) {
			_, file, line, _ := runtime.Caller(0)
			Info("payload")
			return file, line + 1
		}},
		{"package Infof", func(*Logger) (string, int) {
			_, file, line, _ := runtime.Caller(0)
			Infof("%s", "payload")
			return file, line + 1
		}},
		{"package Log", func(*Logger) (string, int) {
			_, file, line, _ := runtime.Caller(0)
			Log(InfoLevel, "payload")
			return file, line + 1
		}},
		{"package Logf", func(*Logger) (string, int) {
			_, file, line, _ := runtime.Caller(0)
			Logf(InfoLevel, "%s", "payload")
			return file, line + 1
		}},
	}
	for _, method := range methods {
		t.Run(method.name, func(t *testing.T) {
			sink := &captureHandler{entries: make(chan *message.Entry, 1)}
			logger, err := New(SetRecordCaller(0), SetWorkerConfigs(NewWorkerConfig(NoneLevel, 8).SetHandler(sink)))
			if err != nil {
				t.Fatal(err)
			}
			SetLogger(logger)
			file, line := method.emit(logger)
			if err := logger.Stop(); err != nil {
				t.Fatal(err)
			}
			entry := <-sink.entries
			if entry.Caller == nil || entry.Caller.File != file || entry.Caller.Line != line {
				t.Fatalf("caller = %+v, want %s:%d", entry.Caller, file, line)
			}
		})
	}
}

func logThroughHelper(logger *Logger) { logger.Info("wrapped") }

func TestCallerSkipIsAppliedAfterLoggingFrames(t *testing.T) {
	sink := &captureHandler{entries: make(chan *message.Entry, 1)}
	logger, err := New(SetRecordCaller(1), SetWorkerConfigs(NewWorkerConfig(NoneLevel, 8).SetHandler(sink)))
	if err != nil {
		t.Fatal(err)
	}
	_, file, line, _ := runtime.Caller(0)
	logThroughHelper(logger)
	if err := logger.Stop(); err != nil {
		t.Fatal(err)
	}
	entry := <-sink.entries
	if entry.Caller == nil || entry.Caller.File != file || entry.Caller.Line != line+1 {
		t.Fatalf("caller = %+v, want %s:%d", entry.Caller, file, line+1)
	}
}

func TestLoggerNameSnapshotsAndFormatterOverrides(t *testing.T) {
	for _, kind := range []string{"text", "json", "xml", "shared custom"} {
		for _, override := range []string{"", "worker"} {
			t.Run(kind+"/"+override, func(t *testing.T) {
				stream := &closeAwareBuffer{}
				base := BaseFormatterConfig{LoggerName: override}
				worker := NewWorkerConfig(NoneLevel, 8).SetStreamHandlerConfig(&StreamHandlerConfig{Streamer: stream})
				switch kind {
				case "text":
					worker.SetTextFormatterConfig(&TextFormatterConfig{BaseFormatterConfig: base, PatternStyle: "%[LoggerName]s|%[Message]v"})
				case "json":
					worker.SetJSONFormatterConfig(&JSONFormatterConfig{BaseFormatterConfig: base})
				case "xml":
					worker.SetXMLFormatterConfig(&XMLFormatterConfig{BaseFormatterConfig: base})
				case "shared custom":
					worker.SetFormatter(formatter.NewTextFormatter(TextFormatterConfig{BaseFormatterConfig: base, PatternStyle: "%[LoggerName]s|%[Message]v"}))
				}
				logger, err := New(SetLoggerName("before"), SetWorkerConfigs(worker))
				if err != nil {
					t.Fatal(err)
				}
				logger.Info("first")
				logger.SetLoggerName("after")
				logger.Info("second")
				logger.SetLoggerName("")
				logger.Info("third")
				if err := logger.Stop(); err != nil {
					t.Fatal(err)
				}
				content, _ := stream.snapshot()
				lines := strings.Split(strings.TrimSuffix(content, "\n"), "\n")
				if len(lines) != 3 {
					t.Fatalf("output = %q, want three records", content)
				}
				for index, expectedName := range []string{"before", "after", ""} {
					if override != "" {
						expectedName = override
					}
					actualName := strings.SplitN(lines[index], "|", 2)[0]
					var record message.Record
					if kind == "json" {
						if err := json.Unmarshal([]byte(lines[index]), &record); err != nil {
							t.Fatal(err)
						}
						actualName = record.Module
					} else if kind == "xml" {
						if err := xml.Unmarshal([]byte(lines[index]), &record); err != nil {
							t.Fatal(err)
						}
						actualName = record.Module
					}
					if actualName != expectedName {
						t.Fatalf("record %d name = %q, want %q", index, actualName, expectedName)
					}
				}
			})
		}
	}
}

func TestConcurrentLoggerNameChanges(t *testing.T) {
	sink := &closeAwareHandler{}
	logger, err := New(SetWorkerConfigs(NewWorkerConfig(NoneLevel, 8).SetHandler(sink)))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = logger.Stop() }()
	var workers sync.WaitGroup
	workers.Add(3)
	go func() {
		defer workers.Done()
		for index := 0; index < 500; index++ {
			logger.SetLoggerName(fmt.Sprintf("name-%d", index))
		}
	}()
	go func() {
		defer workers.Done()
		for index := 0; index < 500; index++ {
			_ = logger.GetLoggerName()
		}
	}()
	go func() {
		defer workers.Done()
		for index := 0; index < 500; index++ {
			logger.Info("message")
		}
	}()
	workers.Wait()
}

func TestSharedFormatterDoesNotShareLoggerNames(t *testing.T) {
	shared := formatter.NewTextFormatter(TextFormatterConfig{PatternStyle: "%[LoggerName]s|%[Message]v"})
	firstStream, secondStream := &closeAwareBuffer{}, &closeAwareBuffer{}
	newLogger := func(name string, stream *closeAwareBuffer) *Logger {
		t.Helper()
		worker := NewWorkerConfig(NoneLevel, 8).
			SetStreamHandlerConfig(&StreamHandlerConfig{Streamer: stream}).SetFormatter(shared)
		logger, err := New(SetLoggerName(name), SetWorkerConfigs(worker))
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = logger.Stop() })
		return logger
	}
	first, second := newLogger("first", firstStream), newLogger("second", secondStream)
	first.Info("one")
	first.SetLoggerName("renamed")
	first.Info("two")
	second.Info("three")
	if err := first.Stop(); err != nil {
		t.Fatal(err)
	}
	if err := second.Stop(); err != nil {
		t.Fatal(err)
	}
	if content, _ := firstStream.snapshot(); content != "first|one\nrenamed|two\n" {
		t.Fatalf("first logger output = %q", content)
	}
	if content, _ := secondStream.snapshot(); content != "second|three\n" {
		t.Fatalf("second logger output = %q", content)
	}
}
