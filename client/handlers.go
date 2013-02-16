package client

// this file contains the basic set of event handlers
// to manage tracking an irc connection etc.

import (
	"strings"
)

// sets up the internal event handlers to do essential IRC protocol things
var intHandlers = map[string]HandlerFunc{
	INIT:   (*Conn).h_init,
	"001":  (*Conn).h_001,
	"433":  (*Conn).h_433,
	"CTCP": (*Conn).h_CTCP,
	"NICK": (*Conn).h_NICK,
	"PING": (*Conn).h_PING,
}

func (conn *Conn) addIntHandlers() {
	for n, h := range intHandlers {
		// internal handlers are essential for the IRC client
		// to function, so we don't save their Removers here
		conn.Handle(n, h)
	}
}

// Password/User/Nick broadcast on connection.
func (conn *Conn) h_init(line *Line) {
	if conn.password != "" {
		conn.Pass(conn.password)
	}
	conn.Nick(conn.Me.Nick)
	conn.User(conn.Me.Ident, conn.Me.Name)
}

// Basic ping/pong handler
func (conn *Conn) h_PING(line *Line) {
	conn.Raw("PONG :" + line.Args[0])
}

// Handler to trigger a "CONNECTED" event on receipt of numeric 001
func (conn *Conn) h_001(line *Line) {
	// we're connected!
	conn.dispatch(&Line{Cmd: CONNECTED})
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
	neu := conn.NewNick(line.Args[1])
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

// Handle PRIVMSGs that trigger Commands
func (conn *Conn) h_PRIVMSG(line *Line) {
	txt := line.Args[1]
	if conn.CommandStripNick && strings.HasPrefix(txt, conn.Me.Nick) {
		// Look for '^${nick}[:;>,-]? '
		l := len(conn.Me.Nick)
		switch txt[l] {
		case ':', ';', '>', ',', '-':
			l++
		}
		if txt[l] == ' ' {
			txt = strings.TrimSpace(txt[l:])
		}
	}
	cmd, l := conn.cmdMatch(txt)
	if cmd == nil {
		return
	}
	if conn.CommandStripPrefix {
		txt = strings.TrimSpace(txt[l:])
	}
	if txt != line.Args[1] {
		line = line.Copy()
		line.Args[1] = txt
	}
	cmd.Execute(conn, line)
}

func (conn *Conn) c_HELP(line *Line) {
	if cmd, _ := conn.cmdMatch(line.Args[1]); cmd != nil {
		conn.Privmsg(line.Args[0], cmd.Help())
	}
}
