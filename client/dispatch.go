package client

import (
	"container/list"
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

type handlerElement struct {
	event   string
	handler Handler
}

type handlerSet struct {
	set map[string]*list.List
	sync.RWMutex
}

func newHandlerSet() *handlerSet {
	return &handlerSet{set: make(map[string]*list.List)}
}

func (hs *handlerSet) add(event string, handler Handler) (*list.Element, Remover) {
	hs.Lock()
	defer hs.Unlock()
	event = strings.ToLower(event)
	l, ok := hs.set[event]
	if !ok {
		l = list.New()
		hs.set[event] = l
	}
	element := l.PushBack(&handlerElement{event, handler})
	return element, RemoverFunc(func() {
		hs.remove(element)
	})
}

func (hs *handlerSet) remove(element *list.Element) {
	hs.Lock()
	defer hs.Unlock()
	h := element.Value.(*handlerElement)
	l, ok := hs.set[h.event]
	if !ok {
		logging.Error("Removing node for unknown event '%s'", h.event)
		return
	}
	l.Remove(element)
	if l.Len() == 0 {
		delete(hs.set, h.event)
	}
}

func (hs *handlerSet) dispatch(conn *Conn, line *Line) {
	hs.RLock()
	defer hs.RUnlock()
	event := strings.ToLower(line.Cmd)
	l, ok := hs.set[event]
	if !ok {
		return
	}

	for e := l.Front(); e != nil; e = e.Next() {
		h := e.Value.(*handlerElement)
		go h.handler.Handle(conn, line)
	}
}

type commandElement struct {
	regex    string
	handler  Handler
	priority int
}

type commandList struct {
	list *list.List
	sync.RWMutex
}

func newCommandList() *commandList {
	return &commandList{list: list.New()}
}

func (cl *commandList) add(regex string, handler Handler, priority int) (element *list.Element, remover Remover) {
	cl.Lock()
	defer cl.Unlock()
	c := &commandElement{
		regex:    regex,
		handler:  handler,
		priority: priority,
	}
	// Check for exact regex matches. This will filter out any repeated SimpleCommands.
	for e := cl.list.Front(); e != nil; e = e.Next() {
		c := e.Value.(*commandElement)
		if c.regex == regex {
			logging.Error("Command prefix '%s' already registered.", regex)
			return
		}
	}
	element = cl.list.PushBack(c)
	remover = RemoverFunc(func() {
		cl.remove(element)
	})
	return
}

func (cl *commandList) remove(element *list.Element) {
	cl.Lock()
	defer cl.Unlock()
	cl.list.Remove(element)
}

// Matches the command with the highest priority.
func (cl *commandList) match(text string) (handler Handler) {
	cl.RLock()
	defer cl.RUnlock()
	maxPriority := math.MinInt32
	text = strings.ToLower(text)
	for e := cl.list.Front(); e != nil; e = e.Next() {
		c := e.Value.(*commandElement)
		if c.priority > maxPriority {
			if regex, error := regexp.Compile(c.regex); error == nil {
				if regex.MatchString(text) {
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
	_, remover := conn.handlers.add(name, handler)
	return remover
}

func (conn *Conn) HandleFunc(name string, handlerFunc HandlerFunc) Remover {
	return conn.Handle(name, handlerFunc)
}

func (conn *Conn) Command(regex string, handler Handler, priority int) Remover {
	_, remover := conn.commands.add(regex, handler, priority)
	return remover
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
	stripHandler := func(conn *Conn, line *Line) {
		text := line.Message()
		if conn.cfg.SimpleCommandStripPrefix {
			text = strings.TrimSpace(text[len(prefix):])
		}
		if text != line.Message() {
			line = line.Copy()
			line.Args[1] = text
		}
		handler.Handle(conn, line)
	}
	return conn.CommandFunc(fmt.Sprintf(SimpleCommandRegex, strings.ToLower(prefix)), stripHandler, math.MaxInt32)
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
	command := conn.commands.match(line.Message())
	if command != nil {
		go command.Handle(conn, line)
	}
}
