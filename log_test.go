package log

import (
	"testing"
)

func testLogs() {
	Debug("May there be enough clouds in your life to make a beautiful sunset.")
	Info("May there be enough clouds in your life to make a beautiful sunset.")
	Error("May there be enough clouds in your life to make a beautiful sunset.")
	Warn("May there be enough clouds in your life to make a beautiful sunset.")
	Print("Just because someone doesn't love you the way you want them to,", " doesn't mean they don't love you with all they have.")
	Debugf("%s alone could waken love!", "Love")
	Infof("%s alone could waken love!", "Love")
	Warnf("%s alone could waken love!", "Love")
	Errorf("%s alone could waken love!", "Love")
	Printf("%s alone could waken love!", "Love")
	Fatal("Just because someone doesn't love you the way you want them to,", " doesn't mean they don't love you with all they have.")
	Panic("Just because someone doesn't love you the way you want them to,", " doesn't mean they don't love you with all they have.")
	Fatalf("%s alone could waken love!", "Love")
	Panicf("%s alone could waken love!", "Love")

}
func TestStdoutLog(t *testing.T) {
	err := InitLog(
		SetLoggerName("glog"),
		SetLoggerLevel(DebugLevel),
	)
	if err != nil {
		t.Fatal(err.Error())
	}
	testLogs()
	t.Log("===> PASS!!!")
}

func TestStdoutLogWithEnableCaller(t *testing.T) {
	err := InitLog(
		SetLoggerName("glog"),
		SetLoggerLevel(DebugLevel),
		SetRecordCaller(0),
	)
	if err != nil {
		t.Fatal(err.Error())
	}
	testLogs()
	t.Log("===> PASS!!!")
}

func TestTextFileLog(t *testing.T) {
	err := InitLog(
		SetLoggerName("glog"),
		SetWorkerConfigs(NewDefaultTextFileWorkerConfig("logs")),
	)
	if err != nil {
		t.Fatal(err.Error())
	}
	testLogs()
	t.Log("===> PASS!!!")
}

func TestJsonFileLog(t *testing.T) {
	err := InitLog(
		SetLoggerName("glog"),
		SetWorkerConfigs(NewDefaultJsonFileWorkerConfig("logs")),
	)
	if err != nil {
		t.Fatal(err.Error())
	}
	testLogs()
	t.Log("===> PASS!!!")
}

func TestCustomConfigLog(t *testing.T) {
	err := InitLog(
		SetLoggerName("glog"),
		SetWorkerConfigs(
			NewWorkerConfig(InfoLevel, 1024).SetFileHandlerConfig(
				NewDefaultFileHandlerConfig("logs").WithFileName("text_log"),
			).
				SetJSONFormatterConfig(
					NewDefaultJSONFormatterConfig().
						WithBaseFormatterConfig(
							NewDefaultBaseFormatterConfig().
								WithEnableHostname().
								WithEnableTimestamp().
								WithEnablePid().
								WithEnableIP(),
						),
				),
		),
	)
	if err != nil {
		t.Fatal(err.Error())
	}
	testLogs()
	t.Log("===> PASS!!!")
}
