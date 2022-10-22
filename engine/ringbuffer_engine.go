package engine

import (
	"github.com/ml444/glog/config"
	"github.com/ml444/glog/handler"
	"github.com/ml444/glog/level"
	"github.com/ml444/glog/message"
	"github.com/ml444/samsara"
	"github.com/ml444/samsara/publish"
	"time"
)

type RingBufferEngine struct {
	s             *samsara.Samsara
	producer      publish.IPublisher
	doneChan      chan bool
	enableReport  bool
	reportLevel   level.LogLevel
	normalHandler handler.IHandler
	reportHandler handler.IHandler

	OnError func(msg *message.Entry, err error)
}

func NewRingBufferEngine() *RingBufferEngine {
	return &RingBufferEngine{
		doneChan:     make(chan bool, 1),
		enableReport: config.GlobalConfig.EnableReport,
		reportLevel:  config.GlobalConfig.ReportLevel,
		OnError:      config.GlobalConfig.OnError,
	}
}

func (e *RingBufferEngine) Init() (err error) {
	e.normalHandler, err = handler.GetNewHandler(config.GlobalConfig.Handler.LogHandlerConfig)
	if err != nil {
		e.doneChan <- true
		return err
	}
	if e.enableReport {
		e.reportHandler, err = handler.GetNewHandler(config.GlobalConfig.Handler.ReportHandlerConfig)
		if err != nil {
			e.doneChan <- true
			return err
		}
	}
	e.s = samsara.NewSamsara(int64(config.GlobalConfig.LoggerCacheSize))
	e.producer = e.s.NewSinglePublisher(samsara.NewPublishStrategy(time.Millisecond))
	return nil
}

func (e *RingBufferEngine) handle(msg interface{}) {
	entity, ok := msg.(*message.Entry)
	if !ok {
		return
	}
	err := e.normalHandler.Emit(entity)
	if err != nil && e.OnError != nil {
		e.OnError(entity, err)
	}
	return
}

func (e *RingBufferEngine) reportHandle(msg interface{}) {
	if !e.enableReport {
		return
	}
	entity, ok := msg.(*message.Entry)
	if !ok {
		return
	}
	err := e.reportHandler.Emit(entity)
	if err != nil && e.OnError != nil {
		e.OnError(entity, err)
	}
	return
}

func (e *RingBufferEngine) Start() error {
	e.s.NewSimpleSubscriber(samsara.NewSubscribeStrategy(100*time.Microsecond), e.handle)
	if e.enableReport {
		e.s.NewSimpleSubscriber(samsara.NewSubscribeStrategy(100*time.Microsecond), e.reportHandle)
	}
	e.s.Start()
	return nil
}

func (e *RingBufferEngine) Stop() (err error) {
	err = e.normalHandler.Close()
	if err != nil {
		println(err)
	}
	if e.enableReport && e.reportHandler != nil {
		err = e.reportHandler.Close()
		if err != nil {
			println(err)
		}
	}
	return nil
}

func (e *RingBufferEngine) Send(msg *message.Entry) {
	err := e.producer.Pub(msg)
	if err != nil && e.OnError != nil {
		e.OnError(msg, err)
	}
}
