package client

import (
	"testing"
	"time"
)

func TestHandlerSet(t *testing.T) {
	hs := handlerSet()
	if len(hs.set) != 0 {
		t.Errorf("New set contains things!")
	}

	callcount := 0
	f := func(c *Conn, l *Line) {
		callcount++
	}

	// Add one
	hn1 := hs.add("ONE", HandlerFunc(f)).(*hNode)
	hl, ok := hs.set["one"]
	if len(hs.set) != 1 || !ok {
		t.Errorf("Set doesn't contain 'one' list after add().")
	}
	if hn1.set != hs || hn1.event != "one" || hn1.prev != nil || hn1.next != nil {
		t.Errorf("First node for 'one' not created correctly")
	}
	if hl.start != hn1 || hl.end != hn1 {
		t.Errorf("Node not added to empty 'one' list correctly.")
	}

	// Add another one...
	hn2 := hs.add("one", HandlerFunc(f)).(*hNode)
	if len(hs.set) != 1 {
		t.Errorf("Set contains more than 'one' list after add().")
	}
	if hn2.set != hs || hn2.event != "one" {
		t.Errorf("Second node for 'one' not created correctly")
	}
	if hn1.prev != nil || hn1.next != hn2 || hn2.prev != hn1 || hn2.next != nil {
		t.Errorf("Nodes for 'one' not linked correctly.")
	}
	if hl.start != hn1 || hl.end != hn2 {
		t.Errorf("Node not appended to 'one' list correctly.")
	}

	// Add a third one!
	hn3 := hs.add("one", HandlerFunc(f)).(*hNode)
	if len(hs.set) != 1 {
		t.Errorf("Set contains more than 'one' list after add().")
	}
	if hn3.set != hs || hn3.event != "one" {
		t.Errorf("Third node for 'one' not created correctly")
	}
	if hn1.prev != nil || hn1.next != hn2 ||
		hn2.prev != hn1 || hn2.next != hn3 ||
		hn3.prev != hn2 || hn3.next != nil {
		t.Errorf("Nodes for 'one' not linked correctly.")
	}
	if hl.start != hn1 || hl.end != hn3 {
		t.Errorf("Node not appended to 'one' list correctly.")
	}

	// And finally a fourth one!
	hn4 := hs.add("one", HandlerFunc(f)).(*hNode)
	if len(hs.set) != 1 {
		t.Errorf("Set contains more than 'one' list after add().")
	}
	if hn4.set != hs || hn4.event != "one" {
		t.Errorf("Fourth node for 'one' not created correctly.")
	}
	if hn1.prev != nil || hn1.next != hn2 ||
		hn2.prev != hn1 || hn2.next != hn3 ||
		hn3.prev != hn2 || hn3.next != hn4 ||
		hn4.prev != hn3 || hn4.next != nil {
		t.Errorf("Nodes for 'one' not linked correctly.")
	}
	if hl.start != hn1 || hl.end != hn4 {
		t.Errorf("Node not appended to 'one' list correctly.")
	}

	// Dispatch should result in 4 additions.
	if callcount != 0 {
		t.Errorf("Something incremented call count before we were expecting it.")
	}
	hs.dispatch(nil, &Line{Cmd:"One"})
	<-time.After(time.Millisecond)
	if callcount != 4 {
		t.Errorf("Our handler wasn't called four times :-(")
	}

	// Remove node 3.
	hn3.Remove()
	if len(hs.set) != 1 {
		t.Errorf("Set list count changed after remove().")
	}
	if hn3.set != nil || hn3.prev != nil || hn3.next != nil {
		t.Errorf("Third node for 'one' not removed correctly.")
	}
	if hn1.prev != nil || hn1.next != hn2 ||
		hn2.prev != hn1 || hn2.next != hn4 ||
		hn4.prev != hn2 || hn4.next != nil {
		t.Errorf("Third node for 'one' not unlinked correctly.")
	}
	if hl.start != hn1 || hl.end != hn4 {
		t.Errorf("Third node for 'one' changed list pointers.")
	}

	// Dispatch should result in 3 additions.
	hs.dispatch(nil, &Line{Cmd:"One"})
	<-time.After(time.Millisecond)
	if callcount != 7 {
		t.Errorf("Our handler wasn't called three times :-(")
	}

	// Remove node 1.
	hs.remove(hn1)
	if len(hs.set) != 1 {
		t.Errorf("Set list count changed after remove().")
	}
	if hn1.set != nil || hn1.prev != nil || hn1.next != nil {
		t.Errorf("First node for 'one' not removed correctly.")
	}
	if hn2.prev != nil || hn2.next != hn4 || hn4.prev != hn2 || hn4.next != nil {
		t.Errorf("First node for 'one' not unlinked correctly.")
	}
	if hl.start != hn2 || hl.end != hn4 {
		t.Errorf("First node for 'one' didn't change list pointers.")
	}

	// Dispatch should result in 2 additions.
	hs.dispatch(nil, &Line{Cmd:"One"})
	<-time.After(time.Millisecond)
	if callcount != 9 {
		t.Errorf("Our handler wasn't called two times :-(")
	}

	// Remove node 4.
	hn4.Remove()
	if len(hs.set) != 1 {
		t.Errorf("Set list count changed after remove().")
	}
	if hn4.set != nil || hn4.prev != nil || hn4.next != nil {
		t.Errorf("Fourth node for 'one' not removed correctly.")
	}
	if hn2.prev != nil || hn2.next != nil {
		t.Errorf("Fourth node for 'one' not unlinked correctly.")
	}
	if hl.start != hn2 || hl.end != hn2 {
		t.Errorf("Fourth node for 'one' didn't change list pointers.")
	}

	// Dispatch should result in 1 addition.
	hs.dispatch(nil, &Line{Cmd:"One"})
	<-time.After(time.Millisecond)
	if callcount != 10 {
		t.Errorf("Our handler wasn't called once :-(")
	}

	// Remove node 2.
	hs.remove(hn2)
	if len(hs.set) != 0 {
		t.Errorf("Removing last node in 'one' didn't remove list.")
	}
	if hn2.set != nil || hn2.prev != nil || hn2.next != nil {
		t.Errorf("Second node for 'one' not removed correctly.")
	}
	if hl.start != nil || hl.end != nil {
		t.Errorf("Second node for 'one' didn't change list pointers.")
	}

	// Dispatch should result in NO additions.
	hs.dispatch(nil, &Line{Cmd:"One"})
	<-time.After(time.Millisecond)
	if callcount != 10 {
		t.Errorf("Our handler was called?")
	}
}

func TestCommandSet(t *testing.T) {
	cs := commandSet()
	if len(cs.set) != 0 {
		t.Errorf("New set contains things!")
	}

	c := &command{
		fn: func(c *Conn, l *Line) {},
		help: "wtf?",
	}

	cn1 := cs.add("ONE", c).(*cNode)
	if _, ok := cs.set["one"]; !ok || cn1.set != cs || cn1.prefix != "one" {
		t.Errorf("Command 'one' not added to set correctly.")
	}

	if fail := cs.add("one", c); fail != nil {
		t.Errorf("Adding a second 'one' command did not fail as expected.")
	}
	
	cn2 := cs.add("One Two", c).(*cNode)
	if _, ok := cs.set["one two"]; !ok || cn2.set != cs || cn2.prefix != "one two" {
		t.Errorf("Command 'one two' not added to set correctly.")
	}

	if c, l := cs.match("foo"); c != nil || l != 0 {
		t.Errorf("Matched 'foo' when we shouldn't.")
	}
	if c, l := cs.match("one"); c.(*cNode) != cn1 || l != 3 {
		t.Errorf("Didn't match 'one' when we should have.")
	}
	if c, l := cs.match ("one two three"); c.(*cNode) != cn2 || l != 7 {
		t.Errorf("Didn't match 'one two' when we should have.")
	}

	cs.remove(cn2)
	if _, ok := cs.set["one two"]; ok || cn2.set != nil {
		t.Errorf("Command 'one two' not removed correctly.")
	}
	if c, l := cs.match ("one two three"); c.(*cNode) != cn1 || l != 3 {
		t.Errorf("Didn't match 'one' when we should have.")
	}
	cn1.Remove()
	if _, ok := cs.set["one"]; ok || cn1.set != nil {
		t.Errorf("Command 'one' not removed correctly.")
	}
	if c, l := cs.match ("one two three"); c != nil || l != 0 {
		t.Errorf("Matched 'one' when we shouldn't have.")
	}
}
