package engines

import (
	"github.com/ml444/glog/config"
	"github.com/ml444/glog/handlers"
	"github.com/ml444/glog/levels"
	"github.com/ml444/glog/message"
)

type ChanEngine struct {
	cfg          *config.Config
	msgHandlers  []handlers.IHandler
	msgChan      chan *message.Entry
	reportChan   chan *message.Entry
	doneChan     chan bool
	enableReport bool
	reportLevel  levels.LogLevel

	OnError func(msg *message.Entry, err error)
}

func NewChanEngine(cfg *config.Config) *ChanEngine {
	return &ChanEngine{
		cfg:          cfg,
		enableReport: cfg.EnableReport,
		reportLevel:  cfg.ReportLevel,
	}
}

func (e *ChanEngine) Start() error {
	handler, err := handlers.GetNewHandler(e.cfg.Handler.LogHandlerConfig)
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
		reportHandler, err = handlers.GetNewHandler(e.cfg.Handler.ReportHandlerConfig)
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
func (e *ChanEngine) Init() error {
	e.msgChan = make(chan *message.Entry, e.cfg.LoggerCacheSize)
	e.reportChan = make(chan *message.Entry, e.cfg.ReportCacheSize)
	e.doneChan = make(chan bool, 1)
	return e.Start()
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
