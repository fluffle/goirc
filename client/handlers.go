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
func (conn *Conn) AddHandler(name string, f IRCHandler) {
	conn.ER.AddHandler(name, NewHandler(f))
}

// Wrap f in an anonymous unboxing function
func NewHandler(f IRCHandler) event.Handler {
	return event.NewHandler(func(ev ...interface{}) {
		f(ev[0].(*Conn), ev[1].(*Line))
	})
}

// sets up the internal event handlers to do essential IRC protocol things
func (conn *Conn) addIntHandlers() {
	conn.AddHandler("001", (*Conn).h_001)
	conn.AddHandler("433", (*Conn).h_433)
	conn.AddHandler("CTCP", (*Conn).h_CTCP)
	conn.AddHandler("NICK", (*Conn).h_NICK)
	conn.AddHandler("PING", (*Conn).h_PING)
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

/******************************************************************************\
 * State tracking handlers below here
\******************************************************************************/

func (conn *Conn) addSTHandlers() {
	conn.AddHandler("JOIN", (*Conn).h_JOIN)
	conn.AddHandler("KICK", (*Conn).h_KICK)
	conn.AddHandler("MODE", (*Conn).h_MODE)
	conn.AddHandler("NICK", (*Conn).h_STNICK)
	conn.AddHandler("PART", (*Conn).h_PART)
	conn.AddHandler("QUIT", (*Conn).h_QUIT)
	conn.AddHandler("TOPIC", (*Conn).h_TOPIC)

	conn.AddHandler("311", (*Conn).h_311)
	conn.AddHandler("324", (*Conn).h_324)
	conn.AddHandler("332", (*Conn).h_332)
	conn.AddHandler("352", (*Conn).h_352)
	conn.AddHandler("353", (*Conn).h_353)
	conn.AddHandler("671", (*Conn).h_671)
}

// Handle NICK messages that need to update the state tracker
func (conn *Conn) h_STNICK(line *Line) {
	// all nicks should be handled the same way, our own included
	conn.ST.ReNick(line.Nick, line.Args[0])
}

// Handle JOINs to channels to maintain state
func (conn *Conn) h_JOIN(line *Line) {
	ch := conn.ST.GetChannel(line.Args[0])
	nk := conn.ST.GetNick(line.Nick)
	if ch == nil {
		// first we've seen of this channel, so should be us joining it
		// NOTE this will also take care of nk == nil && ch == nil
		if nk != conn.Me {
			conn.l.Warn("irc.JOIN(): JOIN to unknown channel %s received "+
				"from (non-me) nick %s", line.Args[0], line.Nick)
			return
		}
		ch = conn.ST.NewChannel(line.Args[0])
		// since we don't know much about this channel, ask server for info
		// we get the channel users automatically in 353 and the channel
		// topic in 332 on join, so we just need to get the modes
		conn.Mode(ch.Name)
		// sending a WHO for the channel is MUCH more efficient than
		// triggering a WHOIS on every nick from the 353 handler
		conn.Who(ch.Name)
	}
	if nk == nil {
		// this is the first we've seen of this nick
		nk = conn.ST.NewNick(line.Nick)
		nk.Ident = line.Ident
		nk.Host = line.Host
		// since we don't know much about this nick, ask server for info
		conn.Who(nk.Nick)
	}
	// this takes care of both nick and channel linking \o/
	conn.ST.Associate(ch, nk)
}

// Handle PARTs from channels to maintain state
func (conn *Conn) h_PART(line *Line) {
	conn.ST.Dissociate(conn.ST.GetChannel(line.Args[0]),
		conn.ST.GetNick(line.Nick))
}

// Handle KICKs from channels to maintain state
func (conn *Conn) h_KICK(line *Line) {
	// XXX: this won't handle autorejoining channels on KICK
	// it's trivial to do this in a seperate handler...
	conn.ST.Dissociate(conn.ST.GetChannel(line.Args[0]),
		conn.ST.GetNick(line.Args[1]))
}

// Handle other people's QUITs
func (conn *Conn) h_QUIT(line *Line) {
	conn.ST.DelNick(line.Nick)
}

// Handle MODE changes for channels we know about (and our nick personally)
func (conn *Conn) h_MODE(line *Line) {
	if ch := conn.ST.GetChannel(line.Args[0]); ch != nil {
		// channel modes first
		ch.ParseModes(line.Args[1], line.Args[2:]...)
	} else if nk := conn.ST.GetNick(line.Args[0]); nk != nil {
		// nick mode change, should be us
		if nk != conn.Me {
			conn.l.Warn("irc.MODE(): recieved MODE %s for (non-me) nick %s",
				line.Args[1], line.Args[0])
			return
		}
		nk.ParseModes(line.Args[1])
	} else {
		conn.l.Warn("irc.MODE(): not sure what to do with MODE %s",
			strings.Join(line.Args, " "))
	}
}

// Handle TOPIC changes for channels
func (conn *Conn) h_TOPIC(line *Line) {
	if ch := conn.ST.GetChannel(line.Args[0]); ch != nil {
		ch.Topic = line.Args[1]
	} else {
		conn.l.Warn("irc.TOPIC(): topic change on unknown channel %s",
			line.Args[0])
	}
}

// Handle 311 whois reply
func (conn *Conn) h_311(line *Line) {
	if nk := conn.ST.GetNick(line.Args[1]); nk != nil {
		nk.Ident = line.Args[2]
		nk.Host = line.Args[3]
		nk.Name = line.Args[5]
	} else {
		conn.l.Warn("irc.311(): received WHOIS info for unknown nick %s",
			line.Args[1])
	}
}

// Handle 324 mode reply
func (conn *Conn) h_324(line *Line) {
	if ch := conn.ST.GetChannel(line.Args[1]); ch != nil {
		ch.ParseModes(line.Args[2], line.Args[3:]...)
	} else {
		conn.l.Warn("irc.324(): received MODE settings for unknown channel %s",
			line.Args[1])
	}
}

// Handle 332 topic reply on join to channel
func (conn *Conn) h_332(line *Line) {
	if ch := conn.ST.GetChannel(line.Args[1]); ch != nil {
		ch.Topic = line.Args[2]
	} else {
		conn.l.Warn("irc.332(): received TOPIC value for unknown channel %s",
			line.Args[1])
	}
}

// Handle 352 who reply
func (conn *Conn) h_352(line *Line) {
	if nk := conn.ST.GetNick(line.Args[5]); nk != nil {
		nk.Ident = line.Args[2]
		nk.Host = line.Args[3]
		// XXX: do we care about the actual server the nick is on?
		//      or the hop count to this server?
		// last arg contains "<hop count> <real name>"
		a := strings.SplitN(line.Args[len(line.Args)-1], " ", 2)
		nk.Name = a[1]
		if idx := strings.Index(line.Args[6], "*"); idx != -1 {
			nk.Modes.Oper = true
		}
		if idx := strings.Index(line.Args[6], "H"); idx != -1 {
			nk.Modes.Invisible = true
		}
	} else {
		conn.l.Warn("irc.352(): received WHO reply for unknown nick %s",
			line.Args[5])
	}
}

// Handle 353 names reply
func (conn *Conn) h_353(line *Line) {
	if ch := conn.ST.GetChannel(line.Args[2]); ch != nil {
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
				nk := conn.ST.GetNick(nick)
				if nk == nil {
					// we don't know this nick yet!
					nk = conn.ST.NewNick(nick)
				}
				cp, ok := conn.ST.IsOn(ch.Name, nick)
				if !ok {
					// This nick isn't associated with this channel yet!
					cp = conn.ST.Associate(ch, nk)
				}
				switch c {
				case '~':
					cp.Owner = true
				case '&':
					cp.Admin = true
				case '@':
					cp.Op = true
				case '%':
					cp.HalfOp = true
				case '+':
					cp.Voice = true
				}
			}
		}
	} else {
		conn.l.Warn("irc.353(): received NAMES list for unknown channel %s",
			line.Args[2])
	}
}

// Handle 671 whois reply (nick connected via SSL)
func (conn *Conn) h_671(line *Line) {
	if nk := conn.ST.GetNick(line.Args[1]); nk != nil {
		nk.Modes.SSL = true
	} else {
		conn.l.Warn("irc.671(): received WHOIS SSL info for unknown nick %s",
			line.Args[1])
	}
}
