package client

import (
	"errors"
	"github.com/fluffle/goirc/logging"
	"reflect"
	"runtime"
	"strings"
	"sync"
	"time"
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
	name       string
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
		name:    runtime.FuncForPC(reflect.ValueOf(h).Pointer()).Name(),
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

func (hs *hSet) dispatch(conn *Conn, line *Line) error {
	hs.RLock()
	defer hs.RUnlock()
	ev := strings.ToLower(line.Cmd)
	list, ok := hs.set[ev]
	if !ok {
		return nil
	}
	wg := &sync.WaitGroup{}
	for hn := list.start; hn != nil; hn = hn.next {
		wg.Add(1)
		go func(hn *hNode) {
			logging.Debug("Starting %s handler for event %s", hn.name, line.Cmd)
			hn.Handle(conn, line.Copy())
			logging.Debug("Finished %s handler for event %s", hn.name, line.Cmd)
			wg.Done()
		}(hn)
	}

	// If we don't care about how long handlers run, wait and bail out early
	if conn.cfg.HandlerTimeout == 0 {
		wg.Wait()
		return nil
	}

	// Limit the amount of time we wait.
	endChan := make(chan struct{})
	go func() {
		defer close(endChan)
		wg.Wait()
	}()
	select {
	case <-endChan:
		return nil
	case <-time.After(conn.cfg.HandlerTimeout):
		msg := "Timeout waiting for handlers to complete"
		logging.Error(msg)
		return errors.New(msg)
	}
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

func (conn *Conn) dispatch(line *Line) error {
	// We run the internal handlers first, including all state tracking ones.
	// This ensures that user-supplied handlers that use the tracker have a
	// consistent view of the connection state in handlers that mutate it.
	err := conn.intHandlers.dispatch(conn, line)
	if err != nil {
		return err
	}
	go conn.bgHandlers.dispatch(conn, line)
	err = conn.fgHandlers.dispatch(conn, line)
	if err != nil {
		return err
	}
	return nil
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
