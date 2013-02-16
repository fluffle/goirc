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
	if !ok { return }
	for hn := list.start; hn != nil; hn = hn.next {
		go hn.Handle(conn, line)
	}
}

// An IRC command looks like this:
type Command interface {
	Execute(*Conn, *Line)
	Help() string
}

type command struct {
	fn   HandlerFunc
	help string
}

func (c *command) Execute(conn *Conn, line *Line) {
	c.fn(conn, line)
}

func (c *command) Help() string {
	return c.help
}

type cNode struct {
	cmd    Command
	set    *cSet
	prefix string
}

func (cn *cNode) Execute(conn *Conn, line *Line) {
	cn.cmd.Execute(conn, line)
}

func (cn *cNode) Help() string {
	return cn.cmd.Help()
}

func (cn *cNode) Remove() {
	cn.set.remove(cn)
}

type cSet struct {
	set map[string]*cNode
	sync.RWMutex
}

func commandSet() *cSet {
	return &cSet{set: make(map[string]*cNode)}
}

func (cs *cSet) add(pf string, c Command) Remover {
	cs.Lock()
	defer cs.Unlock()
	pf = strings.ToLower(pf)
	if _, ok := cs.set[pf]; ok {
		logging.Error("Command prefix '%s' already registered.", pf)
		return nil
	}
	cn := &cNode{
		cmd:    c,
		set:    cs,
		prefix: pf,
	}
	cs.set[pf] = cn
	return cn
}

func (cs *cSet) remove(cn *cNode) {
	cs.Lock()
	defer cs.Unlock()
	delete(cs.set, cn.prefix)
	cn.set = nil
}

func (cs *cSet) match(txt string) (final Command, prefixlen int) {
	cs.RLock()
	defer cs.RUnlock()
	txt = strings.ToLower(txt)
	for prefix, cmd := range cs.set {
		if !strings.HasPrefix(txt, prefix) {
			continue
		}
		if final == nil || len(prefix) > prefixlen {
			prefixlen = len(prefix)
			final = cmd
		}
	}
	return
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

func (conn *Conn) Command(prefix string, c Command) Remover {
	return conn.commands.add(prefix, c)
}

func (conn *Conn) CommandFunc(prefix string, hf HandlerFunc, help string) Remover {
	return conn.Command(prefix, &command{hf, help})
}

func (conn *Conn) dispatch(line *Line) {
	conn.handlers.dispatch(conn, line)
}

func (conn *Conn) cmdMatch(txt string) (Command, int) {
	return conn.commands.match(txt)
}
