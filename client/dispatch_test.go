package client

import (
	"testing"
	"time"
)

func TestHandlerSet(t *testing.T) {
	hs := newHandlerSet()
	if len(hs.set) != 0 {
		t.Errorf("New set contains things!")
	}

	callcount := 0
	f := func(c *Conn, l *Line) {
		callcount++
	}

	// Add one
	hn1 := hs.add("ONE", HandlerFunc(f))
	hl, ok := hs.set["one"]
	if len(hs.set) != 1 || !ok {
		t.Errorf("Set doesn't contain 'one' list after add().")
	}
	if hl.Len() != 1 {
		t.Errorf("List doesn't contain 'one' after add().")
	}

	// Add another one...
	hn2 := hs.add("one", HandlerFunc(f))
	if len(hs.set) != 1 {
		t.Errorf("Set contains more than 'one' list after add().")
	}
	if hl.Len() != 2 {
		t.Errorf("List doesn't contain second 'one' after add().")
	}

	// Add a third one!
	hn3 := hs.add("one", HandlerFunc(f))
	if len(hs.set) != 1 {
		t.Errorf("Set contains more than 'one' list after add().")
	}
	if hl.Len() != 3 {
		t.Errorf("List doesn't contain third 'one' after add().")
	}

	// And finally a fourth one!
	hn4 := hs.add("one", HandlerFunc(f))
	if len(hs.set) != 1 {
		t.Errorf("Set contains more than 'one' list after add().")
	}
	if hl.Len() != 4 {
		t.Errorf("List doesn't contain fourth 'one' after add().")
	}

	// Dispatch should result in 4 additions.
	if callcount != 0 {
		t.Errorf("Something incremented call count before we were expecting it.")
	}
	hs.dispatch(nil, &Line{Cmd: "One"})
	<-time.After(time.Millisecond)
	if callcount != 4 {
		t.Errorf("Our handler wasn't called four times :-(")
	}

	// Remove node 3.
	hn3.Remove()
	if len(hs.set) != 1 {
		t.Errorf("Set list count changed after remove().")
	}
	if hl.Len() != 3 {
		t.Errorf("Third 'one' not removed correctly.")
	}

	// Dispatch should result in 3 additions.
	hs.dispatch(nil, &Line{Cmd: "One"})
	<-time.After(time.Millisecond)
	if callcount != 7 {
		t.Errorf("Our handler wasn't called three times :-(")
	}

	// Remove node 1.
	hn1.Remove()
	if len(hs.set) != 1 {
		t.Errorf("Set list count changed after remove().")
	}
	if hl.Len() != 2 {
		t.Errorf("First 'one' not removed correctly.")
	}

	// Dispatch should result in 2 additions.
	hs.dispatch(nil, &Line{Cmd: "One"})
	<-time.After(time.Millisecond)
	if callcount != 9 {
		t.Errorf("Our handler wasn't called two times :-(")
	}

	// Remove node 4.
	hn4.Remove()
	if len(hs.set) != 1 {
		t.Errorf("Set list count changed after remove().")
	}
	if hl.Len() != 1 {
		t.Errorf("Fourth 'one' not removed correctly.")
	}

	// Dispatch should result in 1 addition.
	hs.dispatch(nil, &Line{Cmd: "One"})
	<-time.After(time.Millisecond)
	if callcount != 10 {
		t.Errorf("Our handler wasn't called once :-(")
	}

	// Remove node 2.
	hn2.Remove()
	if len(hs.set) != 0 {
		t.Errorf("Removing last node in 'one' didn't remove list.")
	}

	// Dispatch should result in NO additions.
	hs.dispatch(nil, &Line{Cmd: "One"})
	<-time.After(time.Millisecond)
	if callcount != 10 {
		t.Errorf("Our handler was called?")
	}
}

func TestCommandSet(t *testing.T) {
	cl := newCommandList()
	if cl.list.Len() != 0 {
		t.Errorf("New list contains things!")
	}

	cn1 := cl.add("one", HandlerFunc(func(c *Conn, l *Line) {}), 0)
	if cl.list.Len() != 1 {
		t.Errorf("Command 'one' not added to list correctly.")
	}

	cn2 := cl.add("one two", HandlerFunc(func(c *Conn, l *Line) {}), 0)
	if cl.list.Len() != 2 {
		t.Errorf("Command 'one two' not added to set correctly.")
	}

	if c := cl.match("foo"); c != nil {
		t.Errorf("Matched 'foo' when we shouldn't.")
	}
	if c := cl.match("one"); c == nil {
		t.Errorf("Didn't match when we should have.")
	}
	if c := cl.match("one two three"); c == nil {
		t.Errorf("Didn't match when we should have.")
	}

	cn2.Remove()
	if cl.list.Len() != 1 {
		t.Errorf("Command 'one two' not removed correctly.")
	}
	if c := cl.match("one two three"); c == nil {
		t.Errorf("Didn't match when we should have.")
	}
	cn1.Remove()
	if cl.list.Len() != 0 {
		t.Errorf("Command 'one two' not removed correctly.")
	}
	if c := cl.match("one two three"); c != nil {
		t.Errorf("Matched 'one' when we shouldn't.")
	}
}
