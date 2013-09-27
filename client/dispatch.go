package client

import (
	"github.com/fluffle/goirc/logging"
	"runtime"
	"strings"
	"sync"
)

// An IRC Handler looks like this:
type Handler interface {
	Handle(*Conn, *Line)
}

// And when they've been added to the client they are removable.
type Remover interface {
	Remove()
}

// A HandlerFunc implements Handler.
type HandlerFunc func(*Conn, *Line)

func (hf HandlerFunc) Handle(conn *Conn, line *Line) {
	hf(conn, line)
}

// Handlers are organised using a map of linked-lists, with each map
// key representing an IRC verb or numeric, and the linked list values
// being handlers that are executed in parallel when a Line from the
// server with that verb or numeric arrives.
type hSet struct {
	set map[string]*hList
	sync.RWMutex
}

type hList struct {
	start, end *hNode
}

// Storing the forward and backward links in the node allows O(1) removal.
// This probably isn't strictly necessary but I think it's kinda nice.
type hNode struct {
	next, prev *hNode
	set        *hSet
	event      string
	handler    Handler
}

// A hNode implements both Handler (with configurable panic recovery)...
func (hn *hNode) Handle(conn *Conn, line *Line) {
	defer conn.cfg.Recover(conn, line)
	hn.handler.Handle(conn, line)
}

// ... and Remover.
func (hn *hNode) Remove() {
	hn.set.remove(hn)
}

func handlerSet() *hSet {
	return &hSet{set: make(map[string]*hList)}
}

// When a new Handler is added for an event, it is wrapped in a hNode and
// returned as a Remover so the caller can remove it at a later time.
func (hs *hSet) add(ev string, h Handler) Remover {
	hs.Lock()
	defer hs.Unlock()
	ev = strings.ToLower(ev)
	l, ok := hs.set[ev]
	if !ok {
		l = &hList{}
	}
	hn := &hNode{
		set:     hs,
		event:   ev,
		handler: h,
	}
	if !ok {
		l.start = hn
	} else {
		hn.prev = l.end
		l.end.next = hn
	}
	l.end = hn
	hs.set[ev] = l
	return hn
}

func (hs *hSet) remove(hn *hNode) {
	hs.Lock()
	defer hs.Unlock()
	l, ok := hs.set[hn.event]
	if !ok {
		logging.Error("Removing node for unknown event '%s'", hn.event)
		return
	}
	if hn.next == nil {
		l.end = hn.prev
	} else {
		hn.next.prev = hn.prev
	}
	if hn.prev == nil {
		l.start = hn.next
	} else {
		hn.prev.next = hn.next
	}
	hn.next = nil
	hn.prev = nil
	hn.set = nil
	if l.start == nil || l.end == nil {
		delete(hs.set, hn.event)
	}
}

func (hs *hSet) dispatch(conn *Conn, line *Line) {
	hs.RLock()
	defer hs.RUnlock()
	ev := strings.ToLower(line.Cmd)
	list, ok := hs.set[ev]
	if !ok {
		return
	}
	wg := &sync.WaitGroup{}
	for hn := list.start; hn != nil; hn = hn.next {
		wg.Add(1)
		go func(hn *hNode) {
			hn.Handle(conn, line.Copy())
			wg.Done()
		}(hn)
	}
	wg.Wait()
}

// Handlers are triggered on incoming Lines from the server, with the handler
// "name" being equivalent to Line.Cmd. Read the RFCs for details on what
// replies could come from the server. They'll generally be things like
// "PRIVMSG", "JOIN", etc. but all the numeric replies are left as ascii
// strings of digits like "332" (mainly because I really didn't feel like
// putting massive constant tables in).
func (conn *Conn) Handle(name string, h Handler) Remover {
	return conn.fgHandlers.add(name, h)
}

func (conn *Conn) HandleBG(name string, h Handler) Remover {
	return conn.bgHandlers.add(name, h)
}

func (conn *Conn) handle(name string, h Handler) Remover {
	return conn.intHandlers.add(name, h)
}

func (conn *Conn) HandleFunc(name string, hf HandlerFunc) Remover {
	return conn.Handle(name, hf)
}

func (conn *Conn) dispatch(line *Line) {
	// We run the internal handlers first, including all state tracking ones.
	// This ensures that user-supplied handlers that use the tracker have a
	// consistent view of the connection state in handlers that mutate it.
	conn.intHandlers.dispatch(conn, line)
	// Background handlers are run in parallel and do not block the event loop.
	// This is useful for things that may need to do significant work.
	go conn.bgHandlers.dispatch(conn, line)
	// Foreground handlers have a guarantee of protocol consistency: all the
	// handlers for one event will have finished before the handlers for the
	// next start processing. They are run in parallel but block the event
	// loop, so care should be taken to ensure these handlers are quick :-)
	conn.fgHandlers.dispatch(conn, line)
}

func (conn *Conn) LogPanic(line *Line) {
	if err := recover(); err != nil {
		_, f, l, _ := runtime.Caller(2)
		logging.Error("%s:%d: panic: %v", f, l, err)
	}
}
