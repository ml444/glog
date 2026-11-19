package errorcallback

import "sync"

type event struct {
	callback func(interface{}, error)
	value    interface{}
	err      error
}

type Dispatcher struct {
	mu      sync.Mutex
	queue   []event
	running bool
}

func (dispatcher *Dispatcher) Report(callback func(interface{}, error), value interface{}, err error) {
	if callback == nil {
		return
	}
	dispatcher.mu.Lock()
	dispatcher.queue = append(dispatcher.queue, event{callback: callback, value: value, err: err})
	if !dispatcher.running {
		dispatcher.running = true
		go dispatcher.run()
	}
	dispatcher.mu.Unlock()
}

func (dispatcher *Dispatcher) run() {
	for {
		dispatcher.mu.Lock()
		if len(dispatcher.queue) == 0 {
			dispatcher.queue = nil
			dispatcher.running = false
			dispatcher.mu.Unlock()
			return
		}
		next := dispatcher.queue[0]
		dispatcher.queue[0] = event{}
		dispatcher.queue = dispatcher.queue[1:]
		dispatcher.mu.Unlock()
		next.callback(next.value, next.err)
	}
}
