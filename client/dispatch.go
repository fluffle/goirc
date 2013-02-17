package client

import (
	"fmt"
	"github.com/fluffle/golog/logging"
	"math"
	"regexp"
	"strings"
	"sync"
)

// An IRC handler looks like this:
type Handler interface {
	Handle(*Conn, *Line)
}

type HandlerFunc func(*Conn, *Line)

func (hf HandlerFunc) Handle(conn *Conn, line *Line) {
	hf(conn, line)
}

// And when they've been added to the client they are removable.
type Remover interface {
	Remove()
}

type RemoverFunc func()

func (r RemoverFunc) Remove() {
	r()
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

type command struct {
	handler  Handler
	set      *commandList
	regex    string
	priority int
}

func (c *command) Handle(conn *Conn, line *Line) {
	c.handler.Handle(conn, line)
}

func (c *command) Remove() {
	c.set.remove(c)
}

type commandList struct {
	set []*command
	sync.RWMutex
}

func newCommandList() *commandList {
	return &commandList{}
}

func (cl *commandList) add(regex string, handler Handler, priority int) Remover {
	cl.Lock()
	defer cl.Unlock()
	c := &command{
		handler:  handler,
		set:      cl,
		regex:    regex,
		priority: priority,
	}
	// Check for exact regex matches. This will filter out any repeated SimpleCommands.
	for _, c := range cl.set {
		if c.regex == regex {
			logging.Error("Command prefix '%s' already registered.", regex)
			return nil
		}
	}
	cl.set = append(cl.set, c)
	return c
}

func (cl *commandList) remove(c *command) {
	cl.Lock()
	defer cl.Unlock()
	for index, value := range cl.set {
		if value == c {
			copy(cl.set[index:], cl.set[index+1:])
			cl.set = cl.set[:len(cl.set)-1]
			c.set = nil
			return
		}
	}
}

// Matches the command with the highest priority.
func (cl *commandList) match(txt string) (handler Handler) {
	cl.RLock()
	defer cl.RUnlock()
	maxPriority := math.MinInt32
	for _, c := range cl.set {
		if c.priority > maxPriority {
			if regex, error := regexp.Compile(c.regex); error == nil {
				if regex.MatchString(txt) {
					maxPriority = c.priority
					handler = c.handler
				}
			}
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
func (conn *Conn) Handle(name string, handler Handler) Remover {
	return conn.handlers.add(name, handler)
}

func (conn *Conn) HandleFunc(name string, handlerFunc HandlerFunc) Remover {
	return conn.Handle(name, handlerFunc)
}

func (conn *Conn) Command(regex string, handler Handler, priority int) Remover {
	return conn.commands.add(regex, handler, priority)
}

func (conn *Conn) CommandFunc(regex string, handlerFunc HandlerFunc, priority int) Remover {
	return conn.Command(regex, handlerFunc, priority)
}

var SimpleCommandRegex string = `^!%v(\s|$)`

// Simple commands are commands that are triggered from a simple prefix
// SimpleCommand("roll" handler)
// !roll
// Because simple commands are simple, they get the highest priority.
func (conn *Conn) SimpleCommand(prefix string, handler Handler) Remover {
	return conn.Command(fmt.Sprintf(SimpleCommandRegex, strings.ToLower(prefix)), handler, math.MaxInt32)
}

func (conn *Conn) SimpleCommandFunc(prefix string, handlerFunc HandlerFunc) Remover {
	return conn.SimpleCommand(prefix, handlerFunc)
}

// This will also register a help command to go along with the simple command itself.
// eg. SimpleCommandHelp("bark", "Bot will bark", handler) will make the following commands:
// !bark
// !help bark
func (conn *Conn) SimpleCommandHelp(prefix string, help string, handler Handler) Remover {
	commandCommand := conn.SimpleCommand(prefix, handler)
	helpCommand := conn.SimpleCommandFunc(fmt.Sprintf("help %v", prefix), HandlerFunc(func(conn *Conn, line *Line) {
		conn.Privmsg(line.Target(), help)
	}))
	return RemoverFunc(func() {
		commandCommand.Remove()
		helpCommand.Remove()
	})
}

func (conn *Conn) SimpleCommandHelpFunc(prefix string, help string, handlerFunc HandlerFunc) Remover {
	return conn.SimpleCommandHelp(prefix, help, handlerFunc)
}

func (conn *Conn) dispatch(line *Line) {
	conn.handlers.dispatch(conn, line)
}

func (conn *Conn) command(line *Line) {
	command := conn.commands.match(strings.ToLower(line.Message()))
	if command != nil {
		go command.Handle(conn, line)
	}

}
