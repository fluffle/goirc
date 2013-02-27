package client

import (
	"github.com/fluffle/golog/logging"
	"strings"
	"sync"
)

// An IRC handler looks like this:
type Handler interface {
	Handle(*Conn, *Line)
}

// And when they've been added to the client they are removable.
type Remover interface {
	Remove()
}

type HandlerFunc func(*Conn, *Line)

func (hf HandlerFunc) Handle(conn *Conn, line *Line) {
	hf(conn, line)
}

type hList struct {
	start, end *hNode
}

type hNode struct {
	next, prev *hNode
	set        *hSet
	event      string
	handler    Handler
}

func (hn *hNode) Handle(conn *Conn, line *Line) {
	hn.handler.Handle(conn, line)
}

func (hn *hNode) Remove() {
	hn.set.remove(hn)
}

type hSet struct {
	set map[string]*hList
	sync.RWMutex
}

func handlerSet() *hSet {
	return &hSet{set: make(map[string]*hList)}
}

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
	for hn := list.start; hn != nil; hn = hn.next {
		go hn.Handle(conn, line)
	}
}

// Handlers are triggered on incoming Lines from the server, with the handler
// "name" being equivalent to Line.Cmd. Read the RFCs for details on what
// replies could come from the server. They'll generally be things like
// "PRIVMSG", "JOIN", etc. but all the numeric replies are left as ascii
// strings of digits like "332" (mainly because I really didn't feel like
// putting massive constant tables in).
func (conn *Conn) Handle(name string, h Handler) Remover {
	return conn.handlers.add(name, h)
}

func (conn *Conn) HandleFunc(name string, hf HandlerFunc) Remover {
	return conn.Handle(name, hf)
}

func (conn *Conn) dispatch(line *Line) {
	conn.handlers.dispatch(conn, line)
}
