package engine

import (
	"github.com/ml444/glog/config"
	"github.com/ml444/glog/handler"
	"github.com/ml444/glog/level"
	"github.com/ml444/glog/message"
)

type ChannelEngine struct {
	msgHandlers  []handler.IHandler
	msgChan      chan *message.Entry
	reportChan   chan *message.Entry
	doneChan     chan bool
	enableReport bool
	reportLevel  level.LogLevel

	OnError func(msg *message.Entry, err error)
}

func NewChannelEngine() *ChannelEngine {
	return &ChannelEngine{
		enableReport: config.GlobalConfig.EnableReport,
		reportLevel:  config.GlobalConfig.ReportLevel,
	}
}

func (e *ChannelEngine) Init() error {
	e.msgChan = make(chan *message.Entry, config.GlobalConfig.LoggerCacheSize)
	e.reportChan = make(chan *message.Entry, config.GlobalConfig.ReportCacheSize)
	e.doneChan = make(chan bool, 1)
	return nil
}

func (e *ChannelEngine) Start() error {
	h, err := handler.GetNewHandler(config.GlobalConfig.Handler.LogHandlerConfig)
	if err != nil {
		e.doneChan <- true
		return err
	}
	e.msgHandlers = append(e.msgHandlers, h)
	go func() {
		for {
			select {
			case msg := <-e.msgChan:
				err = h.Emit(msg)
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
		var reportHandler handler.IHandler
		reportHandler, err = handler.GetNewHandler(config.GlobalConfig.Handler.ReportHandlerConfig)
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

func (e *ChannelEngine) Send(entry *message.Entry) {
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

func (e *ChannelEngine) Stop() (err error) {
	for _, h := range e.msgHandlers {
		err = h.Close()
		if err != nil {
			println(err)
		}
	}
	return nil
}
