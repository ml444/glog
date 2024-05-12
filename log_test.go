package log

import (
	"fmt"
	"testing"
)

func TestLog(t *testing.T) {
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
	Exit()
	t.Log("===> PASS!!!")
}

type AA struct {
	s string
}

func (a AA) String() string {
	return a.s
}
func TestName(t *testing.T) {
	a := AA{s: "is"}
	t.Log(fmt.Sprint("cml", 12, a, "a man"))
	fmt.Println("cml", "is", "a man")
}
