package engine

import (
	"github.com/ml444/glog/config"
	"github.com/ml444/glog/handler"
	"github.com/ml444/glog/level"
	"github.com/ml444/glog/message"
)

type ChannelEngine struct {
	cfg         *config.Config
	msgChan     chan *message.Entry
	reportChan  chan *message.Entry
	msgHandlers []handler.IHandler
	OnError     func(msg *message.Entry, err error)
	reportLevel level.LogLevel

	enableReport bool
	done         bool
}

func NewChannelEngine(cfg *config.Config) *ChannelEngine {
	return &ChannelEngine{
		cfg:          cfg,
		enableReport: cfg.EnableReport,
		reportLevel:  cfg.ReportLevel,
		msgChan:      make(chan *message.Entry, cfg.LoggerCacheSize),
		reportChan:   make(chan *message.Entry, cfg.ReportCacheSize),
		OnError:      cfg.OnError,
	}
}

func (e *ChannelEngine) Start() error {
	h, err := handler.GetNewHandler(e.cfg.LogHandlerConfig)
	if err != nil {
		return err
	}
	e.msgHandlers = append(e.msgHandlers, h)
	go func() {
		for !e.done {
			msg := <-e.msgChan
			err = h.Emit(msg)
			if err != nil && e.OnError != nil {
				e.OnError(msg, err)
			}
		}
	}()
	if e.enableReport {
		var reportHandler handler.IHandler
		reportHandler, err = handler.GetNewHandler(e.cfg.ReportHandlerConfig)
		if err != nil {
			return err
		}
		e.msgHandlers = append(e.msgHandlers, reportHandler)
		go func() {
			for !e.done {
				msg := <-e.reportChan
				err = reportHandler.Emit(msg)
				if err != nil && e.OnError != nil {
					e.OnError(msg, err)
				}
			}
		}()
	}
	return nil
}

func (e *ChannelEngine) Send(entry *message.Entry) {
	if e.done {
		close(e.msgChan)
		if e.enableReport {
			close(e.reportChan)
		}
		return
	}
	e.msgChan <- entry
	if e.enableReport && entry.Level >= e.reportLevel {
		e.reportChan <- entry
	}
}

func (e *ChannelEngine) Stop() (err error) {
	for _, h := range e.msgHandlers {
		err = h.Close()
		if err != nil {
			if e.OnError != nil {
				e.OnError(&message.Entry{}, err)
			} else {
				println(err)
			}
		}
	}
	e.done = true
	return nil
}