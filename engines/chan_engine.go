package engines

import (
	"errors"
	"fmt"
	"github.com/ml444/glog/config"
	"github.com/ml444/glog/handlers"
	"github.com/ml444/glog/levels"
	"github.com/ml444/glog/message"
	"strings"
)


type ChanEngine struct {
	cfg             *config.Config
	msgHandlers     []handlers.IHandler
	msgChan         chan *message.Entry
	warnChan        chan *message.Entry
	reportChan      chan *message.Entry
	doneChan        chan bool
	enableReport    bool
	enableSaveAsErr bool
	warmLevel       levels.LogLevel
}

func NewChanEngine(cfg *config.Config) *ChanEngine {
	return &ChanEngine{
		cfg:             cfg,
		enableReport:    cfg.Engine.EnableReport,
	}
}

func (e *ChanEngine) Start() {
	// 启动常规日志处理程序
	handler, err := handlers.GetNewHandler(e.cfg.Handler.CommonConfig)
	if err != nil {
		e.doneChan <- true
		return
	}
	e.msgHandlers = append(e.msgHandlers, handler)
	go func() {
		for {
			select {
			case msg := <-e.msgChan:
				err = handler.Emit(msg)
				if err != nil {
					println(err)
				}
			case <-e.doneChan:
				e.Stop()
				return
			}
		}
	}()

	// 判断是否启动上报处理程序
	if e.enableReport {
		reportHandler, err := handlers.GetNewHandler(e.cfg.Handler.ReportConfig)
		if err != nil {
			e.doneChan <- true
			return
		}
		e.msgHandlers = append(e.msgHandlers, reportHandler)
		go func() {

			for {
				select {
				case msg := <-e.reportChan:
					err = handler.Emit(msg)
					if err != nil {
						fmt.Printf("err: %v \n", err)
					}
				case <-e.doneChan:
					e.Stop()
					return
				}
			}
		}()
	}

}
func (e *ChanEngine) Init() error {
	//e.maxFileSize = defaultMaxFileSize
	e.msgChan = make(chan *message.Entry, e.cfg.Engine.LogCacheSize)
	if e.enableReport {
		e.reportChan = make(chan *message.Entry, e.cfg.Engine.ReportCacheSize)
	}
	e.doneChan = make(chan bool, 1)
	e.Start()
	return nil
}

func (e *ChanEngine) Send(entry *message.Entry) (err error) {
	select {
	case e.msgChan <- entry:
	// sent
	default:
		//println("waring: buffer channel full")
		err = errors.New("buffer channel full")
	}

	if e.enableReport {
		select {
		case e.reportChan <- entry:
		default:
			//println("waring: buffer channel full")
			err = errors.New("buffer channel full")
		}
	}

	if e.enableSaveAsErr && entry.Level >= e.warmLevel {
		select {
		case e.warnChan <- entry:
		default:
			//println("waring: buffer channel full")
			err = errors.New("buffer channel full")
		}
	}
	return err
}

func (e *ChanEngine) Sync() (err error) {
	for _, h := range e.msgHandlers {
		handler := h
		go func() {
			err2 := handler.Sync()
			if err2 != nil {
				fmt.Printf("err: %v \n", err2)
				err = err2
			}
		}()
	}
	return nil
}
func (e *ChanEngine) Stop() {
	for _, handler := range e.msgHandlers {
		handler.Flush()
	}
}

func removeSuffixIfMatched(s string, suffix string) string {
	if strings.HasSuffix(s, suffix) {
		return s[0 : len(s)-len(suffix)]
	}
	return s
}
