package client

// this file contains the basic set of event handlers
// to manage tracking an irc connection etc.

import (
	"github.com/fluffle/goirc/event"
	"github.com/fluffle/goirc/logging"
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
func (conn *Conn) AddHandler(name string, f IRCHandler) {
	conn.Registry.AddHandler(name, NewHandler(f))
}

// Wrap f in an anonymous unboxing function
func NewHandler(f IRCHandler) event.Handler {
	return event.NewHandler(func(ev ...interface{}) {
		f(ev[0].(*Conn), ev[1].(*Line))
	})
}

// Basic ping/pong handler
func (conn *Conn) h_PING(line *Line) {
	conn.Raw("PONG :" + line.Args[0])
}

// Handler to trigger a "CONNECTED" event on receipt of numeric 001
func (conn *Conn) h_001(line *Line) {
	// we're connected!
	conn.Dispatcher.Dispatch("connected", conn, line)
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
	conn.Nick(line.Args[1] + "_")
	// if this is happening before we're properly connected (i.e. the nick
	// we sent in the initial NICK command is in use) we will not receive
	// a NICK message to confirm our change of nick, so ReNick here...
	if line.Args[1] == conn.Me.Nick {
		conn.Me.ReNick(line.Args[1] + "_")
	}
}

// Handler NICK messages to inform us about nick changes
func (conn *Conn) h_NICK(line *Line) {
	// all nicks should be handled the same way, our own included
	if n := conn.GetNick(line.Nick); n != nil {
		n.ReNick(line.Args[0])
	} else {
		logging.Warn("irc.NICK(): unknown nick %s.", line.Nick)
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

// Handle JOINs to channels to maintain state
func (conn *Conn) h_JOIN(line *Line) {
	ch := conn.GetChannel(line.Args[0])
	n := conn.GetNick(line.Nick)
	if ch == nil {
		// first we've seen of this channel, so should be us joining it
		// NOTE this will also take care of n == nil && ch == nil
		if n != conn.Me {
			logging.Warn("irc.JOIN(): JOIN to unknown channel %s recieved "+
				"from (non-me) nick %s", line.Args[0], line.Nick)
			return
		}
		ch = conn.NewChannel(line.Args[0])
		// since we don't know much about this channel, ask server for info
		// we get the channel users automatically in 353 and the channel
		// topic in 332 on join, so we just need to get the modes
		conn.Mode(ch.Name)
		// sending a WHO for the channel is MUCH more efficient than
		// triggering a WHOIS on every nick from the 353 handler
		conn.Who(ch.Name)
	}
	if n == nil {
		// this is the first we've seen of this nick
		n = conn.NewNick(line.Nick, line.Ident, "", line.Host)
		// since we don't know much about this nick, ask server for info
		conn.Who(n.Nick)
	}
	// this takes care of both nick and channel linking \o/
	ch.AddNick(n)
}

// Handle PARTs from channels to maintain state
func (conn *Conn) h_PART(line *Line) {
	ch := conn.GetChannel(line.Args[0])
	n := conn.GetNick(line.Nick)
	if ch != nil && n != nil {
		if _, ok := ch.Nicks[n]; ok {
			ch.DelNick(n)
		} else {
			logging.Warn("irc.PART(): nick %s is not on channel %s",
				line.Nick, line.Args[0])
		}
	} else {
		logging.Warn("irc.PART(): PART of channel %s by nick %s",
			line.Args[0], line.Nick)
	}
}

// Handle KICKs from channels to maintain state
func (conn *Conn) h_KICK(line *Line) {
	// XXX: this won't handle autorejoining channels on KICK
	// it's trivial to do this in a seperate handler...
	ch := conn.GetChannel(line.Args[0])
	n := conn.GetNick(line.Args[1])
	if ch != nil && n != nil {
		if _, ok := ch.Nicks[n]; ok {
			ch.DelNick(n)
		} else {
			logging.Warn("irc.KICK(): nick %s is not on channel %s",
				line.Nick, line.Args[0])
		}
	} else {
		logging.Warn("irc.KICK(): KICK from channel %s of nick %s",
			line.Args[0], line.Args[1])
	}
}

// Handle other people's QUITs
func (conn *Conn) h_QUIT(line *Line) {
	if n := conn.GetNick(line.Nick); n != nil {
		n.Delete()
	} else {
		logging.Warn("irc.QUIT(): QUIT from unknown nick %s", line.Nick)
	}
}

// Handle MODE changes for channels we know about (and our nick personally)
func (conn *Conn) h_MODE(line *Line) {
	// channel modes first
	if ch := conn.GetChannel(line.Args[0]); ch != nil {
		conn.ParseChannelModes(ch, line.Args[1], line.Args[2:])
	} else if n := conn.GetNick(line.Args[0]); n != nil {
		// nick mode change, should be us
		if n != conn.Me {
			logging.Warn("irc.MODE(): recieved MODE %s for (non-me) nick %s",
				line.Args[0], n.Nick)
			return
		}
		conn.ParseNickModes(n, line.Args[1])
	} else {
		logging.Warn("irc.MODE(): not sure what to do with MODE %s",
			strings.Join(line.Args, " "))
	}
}

// Handle TOPIC changes for channels
func (conn *Conn) h_TOPIC(line *Line) {
	if ch := conn.GetChannel(line.Args[0]); ch != nil {
		ch.Topic = line.Args[1]
	} else {
		logging.Warn("irc.TOPIC(): topic change on unknown channel %s",
			line.Args[0])
	}
}

// Handle 311 whois reply
func (conn *Conn) h_311(line *Line) {
	if n := conn.GetNick(line.Args[1]); n != nil {
		n.Ident = line.Args[2]
		n.Host = line.Args[3]
		n.Name = line.Args[5]
	} else {
		logging.Warn("irc.311(): received WHOIS info for unknown nick %s",
			line.Args[1])
	}
}

// Handle 324 mode reply
func (conn *Conn) h_324(line *Line) {
	if ch := conn.GetChannel(line.Args[1]); ch != nil {
		conn.ParseChannelModes(ch, line.Args[2], line.Args[3:])
	} else {
		logging.Warn("irc.324(): received MODE settings for unknown channel %s",
			line.Args[1])
	}
}

// Handle 332 topic reply on join to channel
func (conn *Conn) h_332(line *Line) {
	if ch := conn.GetChannel(line.Args[1]); ch != nil {
		ch.Topic = line.Args[2]
	} else {
		logging.Warn("irc.332(): received TOPIC value for unknown channel %s",
			line.Args[1])
	}
}

// Handle 352 who reply
func (conn *Conn) h_352(line *Line) {
	if n := conn.GetNick(line.Args[5]); n != nil {
		n.Ident = line.Args[2]
		n.Host = line.Args[3]
		// XXX: do we care about the actual server the nick is on?
		//      or the hop count to this server?
		// last arg contains "<hop count> <real name>"
		a := strings.SplitN(line.Args[len(line.Args)-1], " ", 2)
		n.Name = a[1]
		if idx := strings.Index(line.Args[6], "*"); idx != -1 {
			n.Modes.Oper = true
		}
		if idx := strings.Index(line.Args[6], "H"); idx != -1 {
			n.Modes.Invisible = true
		}
	} else {
		logging.Warn("irc.352(): received WHO reply for unknown nick %s",
			line.Args[5])
	}
}

// Handle 353 names reply
func (conn *Conn) h_353(line *Line) {
	if ch := conn.GetChannel(line.Args[2]); ch != nil {
		nicks := strings.Split(line.Args[len(line.Args)-1], " ")
		for _, nick := range nicks {
			// UnrealIRCd's coders are lazy and leave a trailing space
			if nick == "" {
				continue
			}
			switch c := nick[0]; c {
			case '~', '&', '@', '%', '+':
				nick = nick[1:]
				fallthrough
			default:
				n := conn.GetNick(nick)
				if n == nil {
					// we don't know this nick yet!
					n = conn.NewNick(nick, "", "", "")
				}
				if _, ok := ch.Nicks[n]; !ok {
					// This nick isn't associated with this channel yet!
					ch.AddNick(n)
				}
				p := ch.Nicks[n]
				switch c {
				case '~':
					p.Owner = true
				case '&':
					p.Admin = true
				case '@':
					p.Op = true
				case '%':
					p.HalfOp = true
				case '+':
					p.Voice = true
				}
			}
		}
	} else {
		logging.Warn("irc.353(): received NAMES list for unknown channel %s",
			line.Args[2])
	}
}

// Handle 671 whois reply (nick connected via SSL)
func (conn *Conn) h_671(line *Line) {
	if n := conn.GetNick(line.Args[1]); n != nil {
		n.Modes.SSL = true
	} else {
		logging.Warn("irc.671(): received WHOIS SSL info for unknown nick %s",
			line.Args[1])
	}
}

// sets up the internal event handlers to do useful things with lines
func (conn *Conn) SetupHandlers() {
	conn.AddHandler("CTCP", (*Conn).h_CTCP)
	conn.AddHandler("JOIN", (*Conn).h_JOIN)
	conn.AddHandler("KICK", (*Conn).h_KICK)
	conn.AddHandler("MODE", (*Conn).h_MODE)
	conn.AddHandler("NICK", (*Conn).h_NICK)
	conn.AddHandler("PART", (*Conn).h_PART)
	conn.AddHandler("PING", (*Conn).h_PING)
	conn.AddHandler("QUIT", (*Conn).h_QUIT)
	conn.AddHandler("TOPIC", (*Conn).h_TOPIC)

	conn.AddHandler("001", (*Conn).h_001)
	conn.AddHandler("311", (*Conn).h_311)
	conn.AddHandler("324", (*Conn).h_324)
	conn.AddHandler("332", (*Conn).h_332)
	conn.AddHandler("352", (*Conn).h_352)
	conn.AddHandler("353", (*Conn).h_353)
	conn.AddHandler("433", (*Conn).h_433)
	conn.AddHandler("671", (*Conn).h_671)
}
