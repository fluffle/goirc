package client

import (
	"sync/atomic"
	"testing"
	"time"
)

func TestHandlerSet(t *testing.T) {
	// A Conn is needed here because the previous behaviour of passing nil to
	// hset.dispatch causes a nil pointer dereference with panic recovery.
	c, s := setUp(t)
	defer s.tearDown()

	hs := handlerSet()
	if len(hs.set) != 0 {
		t.Errorf("New set contains things!")
	}

	callcount := new(int32)
	f := func(_ *Conn, _ *Line) {
		atomic.AddInt32(callcount, 1)
	}

	// Add one
	hn1 := hs.add("ONE", HandlerFunc(f)).(*hNode)
	_, ok := hs.set["one"]
	if len(hs.set) != 1 || !ok {
		t.Errorf("Set doesn't contain 'one' list after add().")
	}

	// Add another one...
	hn2 := hs.add("one", HandlerFunc(f)).(*hNode)
	if len(hs.set) != 1 {
		t.Errorf("Set contains more than 'one' list after add().")
	}

	// Add a third one!
	hn3 := hs.add("one", HandlerFunc(f)).(*hNode)
	if len(hs.set) != 1 {
		t.Errorf("Set contains more than 'one' list after add().")
	}

	// And finally a fourth one!
	hn4 := hs.add("one", HandlerFunc(f)).(*hNode)
	if len(hs.set) != 1 {
		t.Errorf("Set contains more than 'one' list after add().")
	}

	// Dispatch should result in 4 additions.
	if atomic.LoadInt32(callcount) != 0 {
		t.Errorf("Something incremented call count before we were expecting it.")
	}
	hs.dispatch(c, &Line{Cmd: "One"})
	<-time.After(time.Millisecond)
	if atomic.LoadInt32(callcount) != 4 {
		t.Errorf("Our handler wasn't called four times :-(")
	}

	// Remove node 3.
	hn3.Remove()
	if len(hs.set) != 1 {
		t.Errorf("Set list count changed after remove().")
	}

	// Dispatch should result in 3 additions.
	hs.dispatch(c, &Line{Cmd: "One"})
	<-time.After(time.Millisecond)
	if atomic.LoadInt32(callcount) != 7 {
		t.Errorf("Our handler wasn't called three times :-(")
	}

	// Remove node 1.
	hn1.Remove()
	if len(hs.set) != 1 {
		t.Errorf("Set list count changed after remove().")
	}

	// Dispatch should result in 2 additions.
	hs.dispatch(c, &Line{Cmd: "One"})
	<-time.After(time.Millisecond)
	if atomic.LoadInt32(callcount) != 9 {
		t.Errorf("Our handler wasn't called two times :-(")
	}

	// Remove node 4.
	hn4.Remove()
	if len(hs.set) != 1 {
		t.Errorf("Set list count changed after remove().")
	}

	// Dispatch should result in 1 addition.
	hs.dispatch(c, &Line{Cmd: "One"})
	<-time.After(time.Millisecond)
	if atomic.LoadInt32(callcount) != 10 {
		t.Errorf("Our handler wasn't called once :-(")
	}

	// Remove node 2.
	hn2.Remove()

	// Dispatch should result in NO additions.
	hs.dispatch(c, &Line{Cmd: "One"})
	<-time.After(time.Millisecond)
	if atomic.LoadInt32(callcount) != 10 {
		t.Errorf("Our handler was called?")
	}
}

func TestPanicRecovery(t *testing.T) {
	c, s := setUp(t)
	defer s.tearDown()

	recovered := callCheck(t)
	c.cfg.Recover = func(conn *Conn, line *Line) {
		if err, ok := recover().(string); ok && err == "panic!" {
			recovered.call()
		}
	}
	c.HandleFunc(PRIVMSG, func(conn *Conn, line *Line) {
		panic("panic!")
	})
	c.in <- ParseLine(":nick!user@host.com PRIVMSG #channel :OH NO PIGEONS")
	recovered.assertWasCalled("Failed to recover panic!")
}
