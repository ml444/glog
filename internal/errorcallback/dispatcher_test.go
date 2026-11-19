package errorcallback

import (
	"errors"
	"testing"
	"time"
)

func TestDispatcherSerializesCallbacksWithoutBlockingReporter(t *testing.T) {
	var dispatcher Dispatcher
	started := make(chan struct{})
	release := make(chan struct{})
	defer close(release)
	results := make(chan int, 33)
	want := errors.New("failure")
	dispatcher.Report(func(interface{}, error) {
		close(started)
		<-release
	}, nil, want)
	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("callback did not start")
	}
	for index := 0; index < 32; index++ {
		dispatcher.Report(func(value interface{}, err error) {
			if err != want {
				results <- -1
				return
			}
			results <- value.(int)
		}, index, want)
	}
	select {
	case <-results:
		t.Fatal("callbacks ran concurrently")
	default:
	}
	release <- struct{}{}
	for index := 0; index < 32; index++ {
		select {
		case actual := <-results:
			if actual != index {
				t.Fatalf("callback value = %d, want %d", actual, index)
			}
		case <-time.After(time.Second):
			t.Fatal("queued callback was lost")
		}
	}
	dispatcher.Report(func(interface{}, error) { results <- 32 }, nil, want)
	select {
	case actual := <-results:
		if actual != 32 {
			t.Fatalf("last callback value = %d", actual)
		}
	case <-time.After(time.Second):
		t.Fatal("dispatcher did not process a later report")
	}
}
