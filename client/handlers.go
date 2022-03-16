package client

// this file contains the basic set of event handlers
// to manage tracking an irc connection etc.

import (
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/fluffle/goirc/logging"
)

// sets up the internal event handlers to do essential IRC protocol things
var intHandlers = map[string]HandlerFunc{
	REGISTER: (*Conn).h_REGISTER,
	"001":    (*Conn).h_001,
	"433":    (*Conn).h_433,
	CTCP:     (*Conn).h_CTCP,
	NICK:     (*Conn).h_NICK,
	PING:     (*Conn).h_PING,
	CAP:      (*Conn).h_CAP,
	"410":    (*Conn).h_410,
}

// set up the ircv3 capabilities supported by this client which will be requested by default to the server.
var defaultCaps = []string{}

func (conn *Conn) addIntHandlers() {
	for n, h := range intHandlers {
		// internal handlers are essential for the IRC client
		// to function, so we don't save their Removers here
		conn.handle(n, h)
	}
}

// Basic ping/pong handler
func (conn *Conn) h_PING(line *Line) {
	conn.Pong(line.Args[0])
}

// Handler for initial registration with server once tcp connection is made.
func (conn *Conn) h_REGISTER(line *Line) {
	if conn.cfg.EnableCapabilityNegotiation {
		conn.Cap(CAP_LS)
	}

	if conn.cfg.Pass != "" {
		conn.Pass(conn.cfg.Pass)
	}
	conn.Nick(conn.cfg.Me.Nick)
	conn.User(conn.cfg.Me.Ident, conn.cfg.Me.Name)
}

func (conn *Conn) getRequestCapabilities() *capSet {
	s := capabilitySet()

	// add capabilites supported by the client
	s.Add(defaultCaps...)

	// add capabilites requested by the user
	s.Add(conn.cfg.Capabilites...)

	return s
}

func (conn *Conn) negotiateCapabilities(supportedCaps []string) {
	conn.supportedCaps.Add(supportedCaps...)

	reqCaps := conn.getRequestCapabilities()
	reqCaps.Intersect(conn.supportedCaps)

	if reqCaps.Size() > 0 {
		conn.Cap(CAP_REQ, reqCaps.Slice()...)
	} else {
		conn.Cap(CAP_END)
	}
}

func (conn *Conn) handleCapAck(caps []string) {
	for _, cap := range caps {
		conn.currCaps.Add(cap)
	}
	conn.Cap(CAP_END)
}

func (conn *Conn) handleCapNak(caps []string) {
	conn.Cap(CAP_END)
}

const (
	CAP_LS  = "LS"
	CAP_REQ = "REQ"
	CAP_ACK = "ACK"
	CAP_NAK = "NAK"
	CAP_END = "END"
)

type capSet struct {
	caps map[string]bool
	mu   sync.RWMutex
}

func capabilitySet() *capSet {
	return &capSet{
		caps: make(map[string]bool),
	}
}

func (c *capSet) Add(caps ...string) {
	c.mu.Lock()
	for _, cap := range caps {
		if strings.HasPrefix(cap, "-") {
			c.caps[cap[1:]] = false
		} else {
			c.caps[cap] = true
		}
	}
	c.mu.Unlock()
}

func (c *capSet) Has(cap string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.caps[cap]
}

// Intersect computes the intersection of two sets.
func (c *capSet) Intersect(other *capSet) {
	c.mu.Lock()

	for cap := range c.caps {
		if !other.Has(cap) {
			delete(c.caps, cap)
		}
	}

	c.mu.Unlock()
}

func (c *capSet) Slice() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	capSlice := make([]string, 0, len(c.caps))
	for cap := range c.caps {
		capSlice = append(capSlice, cap)
	}

	// make output predictable for testing
	sort.Strings(capSlice)
	return capSlice
}

func (c *capSet) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.caps)
}

// This handler is triggered when an invalid cap command is received by the server.
func (conn *Conn) h_410(line *Line) {
	logging.Warn("Invalid cap subcommand: ", line.Args[1])
}

// Handler for capability negotiation commands.
// Note that even if multiple CAP_END commands may be sent to the server during negotiation,
// only the first will be considered.
func (conn *Conn) h_CAP(line *Line) {
	subcommand := line.Args[1]

	caps := strings.Fields(line.Text())
	switch subcommand {
	case CAP_LS:
		conn.negotiateCapabilities(caps)
	case CAP_ACK:
		conn.handleCapAck(caps)
	case CAP_NAK:
		conn.handleCapNak(caps)
	}
}

// Handler to trigger a CONNECTED event on receipt of numeric 001
// :<server> 001 <nick> :Welcome message <nick>!<user>@<host>
func (conn *Conn) h_001(line *Line) {
	// We're connected! Defer this for control flow reasons.
	defer conn.dispatch(&Line{Cmd: CONNECTED, Time: time.Now()})

	// Accept the server's opinion of what our nick actually is
	// and record our ident and hostname (from the server's perspective)
	me, nick, t := conn.Me(), line.Target(), line.Text()
	if idx := strings.LastIndex(t, " "); idx != -1 {
		t = t[idx+1:]
	}
	_, ident, host, ok := parseUserHost(t)

	if me.Nick != nick {
		logging.Warn("Server changed our nick on connect: old=%q new=%q", me.Nick, nick)
	}
	if conn.st != nil {
		if ok {
			conn.st.NickInfo(me.Nick, ident, host, me.Name)
		}
		conn.cfg.Me = conn.st.ReNick(me.Nick, nick)
	} else {
		conn.cfg.Me.Nick = nick
		if ok {
			conn.cfg.Me.Ident = ident
			conn.cfg.Me.Host = host
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
	me := conn.Me()
	neu := conn.cfg.NewNick(line.Args[1])
	conn.Nick(neu)
	if !line.argslen(1) {
		return
	}
	// if this is happening before we're properly connected (i.e. the nick
	// we sent in the initial NICK command is in use) we will not receive
	// a NICK message to confirm our change of nick, so ReNick here...
	if line.Args[1] == me.Nick {
		if conn.st != nil {
			conn.cfg.Me = conn.st.ReNick(me.Nick, neu)
		} else {
			conn.cfg.Me.Nick = neu
		}
	}
}

// Handle VERSION requests and CTCP PING
func (conn *Conn) h_CTCP(line *Line) {
	if line.Args[0] == VERSION {
		conn.CtcpReply(line.Nick, VERSION, conn.cfg.Version)
	} else if line.Args[0] == PING && line.argslen(2) {
		conn.CtcpReply(line.Nick, PING, line.Args[2])
	}
}

// Handle updating our own NICK if we're not using the state tracker
func (conn *Conn) h_NICK(line *Line) {
	if conn.st == nil && line.Nick == conn.cfg.Me.Nick {
		conn.cfg.Me.Nick = line.Args[0]
	}
}
