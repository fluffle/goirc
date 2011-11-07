package client

// this file contains the basic set of event handlers
// to manage tracking an irc connection etc.

import (
	"github.com/fluffle/goirc/event"
	"strings"
)

// An IRC handler looks like this:
type IRCHandler func(*Conn, *Line)

// AddHandler() adds an event handler for a specific IRC command.
//
// Handlers are triggered on incoming Lines from the server, with the handler
// "name" being equivalent to Line.Cmd. Read the RFCs for details on what
// replies could come from the server. They'll generally be things like
// "PRIVMSG", "JOIN", etc. but all the numeric replies are left as ascii
// strings of digits like "332" (mainly because I really didn't feel like
// putting massive constant tables in).
func (conn *Conn) AddHandler(name string, f IRCHandler) event.Handler {
	h := NewHandler(f)
	conn.ER.AddHandler(h, name)
	return h
}

// Wrap f in an anonymous unboxing function
func NewHandler(f IRCHandler) event.Handler {
	return event.NewHandler(func(ev ...interface{}) {
		f(ev[0].(*Conn), ev[1].(*Line))
	})
}

// sets up the internal event handlers to do essential IRC protocol things
var intHandlers map[string]event.Handler
func init() {
	intHandlers = make(map[string]event.Handler)
	intHandlers["001"] = NewHandler((*Conn).h_001)
	intHandlers["433"] = NewHandler((*Conn).h_433)
	intHandlers["CTCP"] = NewHandler((*Conn).h_CTCP)
	intHandlers["NICK"] = NewHandler((*Conn).h_NICK)
	intHandlers["PING"] = NewHandler((*Conn).h_PING)
}

func (conn *Conn) addIntHandlers() {
	for n, h := range intHandlers {
		conn.ER.AddHandler(h, n)
	}
}

// Basic ping/pong handler
func (conn *Conn) h_PING(line *Line) {
	conn.Raw("PONG :" + line.Args[0])
}

// Handler to trigger a "CONNECTED" event on receipt of numeric 001
func (conn *Conn) h_001(line *Line) {
	// we're connected!
	conn.ED.Dispatch("connected", conn, line)
	// and we're being given our hostname (from the server's perspective)
	t := line.Args[len(line.Args)-1]
	if idx := strings.LastIndex(t, " "); idx != -1 {
		t = t[idx+1:]
		if idx = strings.Index(t, "@"); idx != -1 {
			conn.Me.Host = t[idx+1:]
		}
	}
}

// XXX: do we need 005 protocol support message parsing here?
// probably in the future, but I can't quite be arsed yet.
/*
	:irc.pl0rt.org 005 GoTest CMDS=KNOCK,MAP,DCCALLOW,USERIP UHNAMES NAMESX SAFELIST HCN MAXCHANNELS=20 CHANLIMIT=#:20 MAXLIST=b:60,e:60,I:60 NICKLEN=30 CHANNELLEN=32 TOPICLEN=307 KICKLEN=307 AWAYLEN=307 :are supported by this server
	:irc.pl0rt.org 005 GoTest MAXTARGETS=20 WALLCHOPS WATCH=128 WATCHOPTS=A SILENCE=15 MODES=12 CHANTYPES=# PREFIX=(qaohv)~&@%+ CHANMODES=beI,kfL,lj,psmntirRcOAQKVCuzNSMT NETWORK=bb101.net CASEMAPPING=ascii EXTBAN=~,cqnr ELIST=MNUCT :are supported by this server
	:irc.pl0rt.org 005 GoTest STATUSMSG=~&@%+ EXCEPTS INVEX :are supported by this server
*/

// Handler to deal with "433 :Nickname already in use"
func (conn *Conn) h_433(line *Line) {
	// Args[1] is the new nick we were attempting to acquire
	neu := line.Args[1] + "_"
	conn.Nick(neu)
	// if this is happening before we're properly connected (i.e. the nick
	// we sent in the initial NICK command is in use) we will not receive
	// a NICK message to confirm our change of nick, so ReNick here...
	if line.Args[1] == conn.Me.Nick {
		if conn.st {
			conn.ST.ReNick(conn.Me.Nick, neu)
		} else {
			conn.Me.Nick = neu
		}
	}
}

// Handle VERSION requests and CTCP PING
func (conn *Conn) h_CTCP(line *Line) {
	if line.Args[0] == "VERSION" {
		conn.CtcpReply(line.Nick, "VERSION", "powered by goirc...")
	} else if line.Args[0] == "PING" {
		conn.CtcpReply(line.Nick, "PING", line.Args[2])
	}
}

// Handle updating our own NICK if we're not using the state tracker
func (conn *Conn) h_NICK(line *Line) {
	if !conn.st && line.Nick == conn.Me.Nick {
		conn.Me.Nick = line.Args[0]
	}
}
