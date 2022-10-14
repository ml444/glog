package engines

import (
	"github.com/ml444/glog/config"
	"github.com/ml444/glog/handlers"
	"github.com/ml444/glog/levels"
	"github.com/ml444/glog/message"
)

type ChanEngine struct {
	msgHandlers  []handlers.IHandler
	msgChan      chan *message.Entry
	reportChan   chan *message.Entry
	doneChan     chan bool
	enableReport bool
	reportLevel  levels.LogLevel

	OnError func(msg *message.Entry, err error)
}

func NewChanEngine() *ChanEngine {
	return &ChanEngine{
		enableReport: config.GlobalConfig.EnableReport,
		reportLevel:  config.GlobalConfig.ReportLevel,
	}
}

func (e *ChanEngine) Init() error {
	e.msgChan = make(chan *message.Entry, config.GlobalConfig.LoggerCacheSize)
	e.reportChan = make(chan *message.Entry, config.GlobalConfig.ReportCacheSize)
	e.doneChan = make(chan bool, 1)
	return nil
}

func (e *ChanEngine) Start() error {
	handler, err := handlers.GetNewHandler(config.GlobalConfig.Handler.LogHandlerConfig)
	if err != nil {
		e.doneChan <- true
		return err
	}
	e.msgHandlers = append(e.msgHandlers, handler)
	go func() {
		for {
			select {
			case msg := <-e.msgChan:
				err = handler.Emit(msg)
				if err != nil && e.OnError != nil {
					e.OnError(msg, err)
				}
			case <-e.doneChan:
				err = e.Stop()
				if err != nil && e.OnError != nil {
					e.OnError(&message.Entry{}, err)
				}
				return
			}
		}
	}()
	if e.enableReport {
		var reportHandler handlers.IHandler
		reportHandler, err = handlers.GetNewHandler(config.GlobalConfig.Handler.ReportHandlerConfig)
		if err != nil {
			e.doneChan <- true
			return err
		}
		e.msgHandlers = append(e.msgHandlers, reportHandler)
		go func() {
			for {
				select {
				case msg := <-e.reportChan:
					err = reportHandler.Emit(msg)
					if err != nil && e.OnError != nil {
						e.OnError(msg, err)
					}
				case <-e.doneChan:
					err = e.Stop()
					if err != nil && e.OnError != nil {
						e.OnError(&message.Entry{}, err)
					}
					return
				}
			}
		}()
	}
	return nil
}

func (e *ChanEngine) Send(entry *message.Entry) {
	select {
	case e.msgChan <- entry:
	}

	if e.enableReport && entry.Level >= e.reportLevel {
		select {
		case e.reportChan <- entry:
		}
	}
	return
}

func (e *ChanEngine) Stop() (err error) {
	for _, h := range e.msgHandlers {
		err = h.Close()
		if err != nil {
			println(err)
		}
	}
	return nil
}
