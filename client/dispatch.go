package client

import (
	"github.com/fluffle/goirc/logging"
	"runtime"
	"strings"
	"sync"
)

// Handlers are triggered on incoming Lines from the server, with the handler
// "name" being equivalent to Line.Cmd. Read the RFCs for details on what
// replies could come from the server. They'll generally be things like
// "PRIVMSG", "JOIN", etc. but all the numeric replies are left as ascii
// strings of digits like "332" (mainly because I really didn't feel like
// putting massive constant tables in).
//
// Foreground handlers have a guarantee of protocol consistency: all the
// handlers for one event will have finished before the handlers for the
// next start processing. They are run in parallel but block the event
// loop, so care should be taken to ensure these handlers are quick :-)
//
// Background handlers are run in parallel and do not block the event loop.
// This is useful for things that may need to do significant work.
type Handler interface {
	Handle(*Conn, *Line)
}

// Removers allow for a handler that has been previously added to the client
// to be removed.
type Remover interface {
	Remove()
}

// HandlerFunc allows a bare function with this signature to implement the
// Handler interface. It is used by Conn.HandleFunc.
type HandlerFunc func(*Conn, *Line)

func (hf HandlerFunc) Handle(conn *Conn, line *Line) {
	hf(conn, line)
}

// Handlers are organised using a map of lockless singly linked lists, with
// each map key representing an IRC verb or numeric, and the linked list
// values being handlers that are executed in parallel when a Line from the
// server with that verb or numeric arrives.
type hSet struct {
	set map[string]*hList
	sync.RWMutex
}

func (hs *hSet) getList(ev string) (hl *hList, ok bool) {
	ev = strings.ToLower(ev)
	hs.RLock()
	defer hs.RUnlock()
	hl, ok = hs.set[ev]
	return
}

func (hs *hSet) getOrMakeList(ev string) (hl *hList) {
	ev = strings.ToLower(ev)
	hs.Lock()
	defer hs.Unlock()
	hl, ok := hs.set[ev]
	if !ok {
		hl = makeHList()
		hs.set[ev] = hl
	}
	return hl
}

// Lists are lockless thanks to atomic pointers. (which hNodePtr wraps)
type hList struct {
	first, last hNodePtr
}

// In order for the whole thing to be goroutine-safe, each list must contain a
// zero-valued node at any given time as its last element.  You'll see why
// later down.
func makeHList() (hl *hList) {
	hl, hn0 := &hList{}, &hNode{}
	hl.first.store(hn0)
	hl.last.store(hn0)
	return
}

// (hNodeState is also an atomic wrapper.)
type hNode struct {
	next    hNodePtr
	state   hNodeState
	handler Handler
}

// Nodes progress through these three stages in order as the program runs.
const (
	unready hNodeState = iota
	active
	unlinkable
)

// A hNode implements both Handler (with configurable panic recovery)...
func (hn *hNode) Handle(conn *Conn, line *Line) {
	defer conn.cfg.Recover(conn, line)
	hn.handler.Handle(conn, line)
}

// ... and Remover, which works by flagging the node so the goroutines running
// hSet.dispatch know to ignore its handler and to dispose of it.
func (hn *hNode) Remove() {
	hn.state.store(unlinkable)
}

func handlerSet() *hSet {
	return &hSet{set: make(map[string]*hList)}
}

// When a new Handler is added for an event, it is assigned into a hNode,
// which is returned as a Remover so the caller can remove it at a later time.
//
// Concerning goroutine-safety, the point is that the atomic swap there
// reserves the previous last node for this handler and puts up a new one.
// The former node has the desirable property that the rest of the list points
// to it, and the latter inherits this property once the former becomes part
// of the list.  It's also the case that handler should't be read by
// hSet.dispatch before the node is marked as ready via state.
func (hs *hSet) add(ev string, h Handler) Remover {
	hl := hs.getOrMakeList(ev)
	hn0 := &hNode{}
	hn := hl.last.swap(hn0)
	hn.next.store(hn0)
	hn.handler = h
	hn.state.compareAndSwap(unready, active)
	return hn
}

// And finally, dispatch works like so: it goes through the whole list while
// remembering the adress of the pointer that led it to the current node,
// which allows it to unlink it if it must be.  Since the pointers are atomic,
// if many goroutine enter the same unlinkable node at the same time, they
// will all end up writing the same value to the pointer anyway.  Even in
// cases where consecutive nodes are flagged and unlinking node n revives node
// n+1 which had been unlinked by making node n point to n+2 without the
// unlinker of n+1 noticing, all dead nodes are unmistakable and will
// eventually be definitely unlinked and garbage-collected.  Also note that
// the fact that the last node is always a zero node, as well as letting the
// list grow concurrently, allows the next-to-last node to be unlinked safely.
func (hs *hSet) dispatch(conn *Conn, line *Line) {
	hl, ok := hs.getList(line.Cmd)
	if !ok {
		return // nothing to do
	}
	wg := &sync.WaitGroup{}
	hn, hnptr := hl.first.load(), &hl.first
	for hn != nil {
		switch hn.state.load() {
		case active:
			wg.Add(1)
			go func(hn *hNode) {
				hn.Handle(conn, line.Copy())
				wg.Done()
			}(hn)
			fallthrough
		case unready:
			hnptr = &hn.next
			hn = hnptr.load()
		case unlinkable:
			hn = hn.next.load()
			hnptr.store(hn)
		}
	}
	wg.Wait()
}

// Handle adds the provided handler to the foreground set for the named event.
// It will return a Remover that allows that handler to be removed again.
func (conn *Conn) Handle(name string, h Handler) Remover {
	return conn.fgHandlers.add(name, h)
}

// HandleBG adds the provided handler to the background set for the named
// event. It may go away in the future.
// It will return a Remover that allows that handler to be removed again.
func (conn *Conn) HandleBG(name string, h Handler) Remover {
	return conn.bgHandlers.add(name, h)
}

func (conn *Conn) handle(name string, h Handler) Remover {
	return conn.intHandlers.add(name, h)
}

// HandleFunc adds the provided function as a handler in the foreground set
// for the named event.
// It will return a Remover that allows that handler to be removed again.
func (conn *Conn) HandleFunc(name string, hf HandlerFunc) Remover {
	return conn.Handle(name, hf)
}

func (conn *Conn) dispatch(line *Line) {
	// We run the internal handlers first, including all state tracking ones.
	// This ensures that user-supplied handlers that use the tracker have a
	// consistent view of the connection state in handlers that mutate it.
	conn.intHandlers.dispatch(conn, line)
	go conn.bgHandlers.dispatch(conn, line)
	conn.fgHandlers.dispatch(conn, line)
}

// LogPanic is used as the default panic catcher for the client. If, like me,
// you are not good with computer, and you'd prefer your bot not to vanish into
// the ether whenever you make unfortunate programming mistakes, you may find
// this useful: it will recover panics from handler code and log the errors.
func (conn *Conn) LogPanic(line *Line) {
	if err := recover(); err != nil {
		_, f, l, _ := runtime.Caller(2)
		logging.Error("%s:%d: panic: %v", f, l, err)
	}
}
