package log

import (
	"io"
	"reflect"
	"sync"

	"github.com/ml444/glog/handler"
)

var (
	resourceMu     sync.Mutex
	resourceOwners = make(map[interface{}]*resourceOwner)
)

type resourceOwner struct {
	closer io.Closer
	refs   int
}

type closerLease struct {
	key      interface{}
	owner    *resourceOwner
	once     sync.Once
	closeErr error
}

func retainCloser(closer io.Closer) *closerLease {
	var key interface{} = closer
	if !reflect.TypeOf(closer).Comparable() {
		key = reflect.ValueOf(closer)
	}
	resourceMu.Lock()
	owner := resourceOwners[key]
	if owner == nil {
		owner = &resourceOwner{closer: closer}
		resourceOwners[key] = owner
	}
	owner.refs++
	resourceMu.Unlock()
	return &closerLease{key: key, owner: owner}
}

func (lease *closerLease) Close() error {
	lease.once.Do(func() {
		resourceMu.Lock()
		lease.owner.refs--
		lastOwner := lease.owner.refs == 0
		if lastOwner {
			delete(resourceOwners, lease.key)
		}
		resourceMu.Unlock()
		if lastOwner {
			lease.closeErr = lease.owner.closer.Close()
		}
	})
	return lease.closeErr
}

type retainedHandler struct {
	handler.IHandler
	lease *closerLease
}

func (h *retainedHandler) Close() error {
	return h.lease.Close()
}

func (h *retainedHandler) BackpressureStats() BackpressureStats {
	if provider, ok := h.IHandler.(handler.BackpressureStatsProvider); ok {
		return provider.BackpressureStats()
	}
	return BackpressureStats{}
}

type retainedStream struct {
	handler.IStreamer
	lease *closerLease
}

func (stream *retainedStream) Close() error {
	return stream.lease.Close()
}
