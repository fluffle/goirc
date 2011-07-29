package event

// oh hey unit tests. or functionality tests, or something.

import (
	"testing"
	"time"
)

func TestSimpleDispatch(t *testing.T) {
	r := NewRegistry()
	out := make(chan bool)
	
	h := NewHandler(func(ev ...interface{}) {
		out <- ev[0].(bool)
	})
	r.AddHandler("send", h)

	r.Dispatch("send", true)
	if val := <-out; !val {
		t.Fail()
	}

	r.Dispatch("send", false)
	if val := <-out; val {
		t.Fail()
	}
}

func TestParallelDispatch(t *testing.T) {
	r := NewRegistry()
	// ensure we have enough of a buffer that all sends complete
	out := make(chan int, 5)
	// handler factory :-)
	factory := func(t int) Handler {
		return NewHandler(func(ev ...interface{}) {
			// t * 10ms sleep
			time.Sleep(int64(t * 1e7))
			out <- t
		})
	}

	// create some handlers and send an event to them
	for _, t := range []int{5,11,2,15,8} {
		r.AddHandler("send", factory(t))
	}
	r.Dispatch("send")

	// If parallel dispatch is working, results from out should be in numerical order
	if val := <-out; val != 2 {
		t.Fail()
	}
	if val := <-out; val != 5 {
		t.Fail()
	}
	if val := <-out; val != 8 {
		t.Fail()
	}
	if val := <-out; val != 11 {
		t.Fail()
	}
	if val := <-out; val != 15 {
		t.Fail()
	}
}

func TestSerialDispatch(t *testing.T) {
	r := NewRegistry()
	r.(*registry).Serial()
	// ensure we have enough of a buffer that all sends complete
	out := make(chan int, 5)
	// handler factory :-)
	factory := func(t int) Handler {
		return NewHandler(func(ev ...interface{}) {
			// t * 10ms sleep
			time.Sleep(int64(t * 1e7))
			out <- t
		})
	}

	// create some handlers and send an event to them
	for _, t := range []int{5,11,2,15,8} {
		r.AddHandler("send", factory(t))
	}
	r.Dispatch("send")

	// If serial dispatch is working, results from out should be in handler order
	if val := <-out; val != 5 {
		t.Fail()
	}
	if val := <-out; val != 11 {
		t.Fail()
	}
	if val := <-out; val != 2 {
		t.Fail()
	}
	if val := <-out; val != 15 {
		t.Fail()
	}
	if val := <-out; val != 8 {
		t.Fail()
	}
}

