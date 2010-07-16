package irc

// this file contains the basic set of event handlers
// to manage tracking an irc connection etc.

import (
	"strings"
	"strconv"
)

// AddHandler() adds an event handler for a specific IRC command.
//
// Handlers take the form of an anonymous function (currently):
//	func(conn *irc.Conn, line *irc.Line) {
//		// handler code here
//	}
//
// Handlers are triggered on incoming Lines from the server, with the handler
// "name" being equivalent to Line.Cmd. Read the RFCs for details on what
// replies could come from the server. They'll generally be things like
// "PRIVMSG", "JOIN", etc. but all the numeric replies are left as ascii
// strings of digits like "332" (mainly because I really didn't feel like 
// putting massive constant tables in).
func (conn *Conn) AddHandler(name string, f func(*Conn, *Line)) {
	n := strings.ToUpper(name)
	if e, ok := conn.events[n]; ok {
		if len(e) == cap(e) {
			// crap, we're full. expand e by another 10 handler slots
			ne := make([]func(*Conn, *Line), len(e), len(e)+10)
			for i := 0; i < len(e); i++ {
				ne[i] = e[i]
			}
			e = ne
		}
		e = e[0 : len(e)+1]
		e[len(e)-1] = f
	} else {
		e := make([]func(*Conn, *Line), 1, 10)
		e[0] = f
		conn.events[n] = e
	}
}

// loops through all event handlers for line.Cmd, running each in a goroutine
func (conn *Conn) dispatchEvent(line *Line) {
	// seems that we end up dispatching an event with a nil line when receiving
	// EOF from the server. Until i've tracked down why....
	if line == nil {
		conn.error("irc.dispatchEvent(): buh? line == nil :-(")
		return
	}

	// So, I think CTCP and (in particular) CTCP ACTION are better handled as
	// separate events as opposed to forcing people to have gargantuan PRIVMSG
	// handlers to cope with the possibilities.
	if line.Cmd == "PRIVMSG" && len(line.Text) > 2 &&
		line.Text[0] == '\001' && line.Text[len(line.Text)-1] == '\001' {
		// WOO, it's a CTCP message
		t := strings.Split(line.Text[1:len(line.Text)-1], " ", 2)
		if c := strings.ToUpper(t[0]); c == "ACTION" {
			// make a CTCP ACTION it's own event a-la PRIVMSG
			line.Cmd = c
		} else {
			// otherwise, dispatch a generic CTCP event that
			// contains the type of CTCP in line.Args[0]
			line.Cmd = "CTCP"
			a := make([]string, len(line.Args)+1)
			a[0] = c
			for i := 0; i < len(line.Args); i++ {
				a[i+1] = line.Args[i]
			}
			line.Args = a
		}
		if len(t) > 1 {
			// for some CTCP messages this could make more sense
			// in line.Args[], but meh. MEH, I say.
			line.Text = t[1]
		}
	}
	if funcs, ok := conn.events[line.Cmd]; ok {
		for _, f := range funcs {
			go f(conn, line)
		}
	}
}

// sets up the internal event handlers to do useful things with lines
// XXX: is there a better way of doing this?
// Turns out there may be but it's not actually implemented in the language yet
// according to the people on freenode/#go-nuts ... :-(
// see: http://golang.org/doc/go_spec.html#Method_expressions for details
// I think this means we should be able to do something along the lines of:
//   conn.AddHandler("event", (*Conn).h_handler);
// where h_handler is declared in the irc package as:
//   func (conn *Conn) h_handler(line *Line) {}
// in the future, but for now the compiler throws a hissy fit.
func (conn *Conn) setupEvents() {
	conn.events = make(map[string][]func(*Conn, *Line))

	// Basic ping/pong handler
	conn.AddHandler("PING", func(conn *Conn, line *Line) { conn.Raw("PONG :" + line.Text) })

	// Handler to trigger a "CONNECTED" event on receipt of numeric 001
	conn.AddHandler("001", func(conn *Conn, line *Line) {
		// we're connected!
		conn.connected = true
		conn.dispatchEvent(&Line{Cmd: "CONNECTED"})
		// and we're being given our hostname (from the server's perspective)
		if ridx := strings.LastIndex(line.Text, " "); ridx != -1 {
			h := line.Text[ridx+1 : len(line.Text)]
			if idx := strings.Index(h, "@"); idx != -1 {
				conn.Me.Host = h[idx+1 : len(h)]
			}
		}
	})

	// XXX: do we need 005 protocol support message parsing here?
	// probably in the future, but I can't quite be arsed yet.
	/*
	:irc.pl0rt.org 005 GoTest CMDS=KNOCK,MAP,DCCALLOW,USERIP UHNAMES NAMESX SAFELIST HCN MAXCHANNELS=20 CHANLIMIT=#:20 MAXLIST=b:60,e:60,I:60 NICKLEN=30 CHANNELLEN=32 TOPICLEN=307 KICKLEN=307 AWAYLEN=307 :are supported by this server
	:irc.pl0rt.org 005 GoTest MAXTARGETS=20 WALLCHOPS WATCH=128 WATCHOPTS=A SILENCE=15 MODES=12 CHANTYPES=# PREFIX=(qaohv)~&@%+ CHANMODES=beI,kfL,lj,psmntirRcOAQKVCuzNSMT NETWORK=bb101.net CASEMAPPING=ascii EXTBAN=~,cqnr ELIST=MNUCT :are supported by this server
	:irc.pl0rt.org 005 GoTest STATUSMSG=~&@%+ EXCEPTS INVEX :are supported by this server
	*/

	// Handler to deal with "433 :Nickname already in use"
	conn.AddHandler("433", func(conn *Conn, line *Line) {
		// Args[1] is the new nick we were attempting to acquire
		conn.Nick(line.Args[1] + "_")
		// if this is happening before we're properly connected (i.e. the nick
		// we sent in the initial NICK command is in use) we will not receive
		// a NICK message to confirm our change of nick, so ReNick here...
		if !conn.connected && line.Args[1] == conn.Me.Nick {
			conn.Me.ReNick(line.Args[1] + "_")
		}
	})

	// Handler NICK messages to inform us about nick changes
	conn.AddHandler("NICK", func(conn *Conn, line *Line) {
		// all nicks should be handled the same way, our own included
		if n := conn.GetNick(line.Nick); n != nil {
			n.ReNick(line.Text)
		} else {
			conn.error("irc.NICK(): buh? unknown nick %s.", line.Nick)
		}
	})

	// Handle VERSION requests and CTCP PING
	conn.AddHandler("CTCP", func(conn *Conn, line *Line) {
		if line.Args[0] == "VERSION" {
			conn.CtcpReply(line.Nick, "VERSION", "powered by goirc...")
		} else if line.Args[0] == "PING" {
			conn.CtcpReply(line.Nick, "PING", line.Text)
		}
	})

	// Handle JOINs to channels to maintain state
	conn.AddHandler("JOIN", func(conn *Conn, line *Line) {
		ch := conn.GetChannel(line.Text)
		n := conn.GetNick(line.Nick)
		if ch == nil {
			// first we've seen of this channel, so should be us joining it
			// NOTE this will also take care of n == nil && ch == nil
			if n != conn.Me {
				conn.error("irc.JOIN(): buh? JOIN to unknown channel %s recieved from (non-me) nick %s", line.Text, line.Nick)
				return
			}
			ch = conn.NewChannel(line.Text)
			// since we don't know much about this channel, ask server for info
			// we get the channel users automatically in 353 and the channel
			// topic in 332 on join, so we just need to get the modes
			conn.Mode(ch.Name,"")
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
	})

	// Handle PARTs from channels to maintain state
	conn.AddHandler("PART", func(conn *Conn, line *Line) {
		ch := conn.GetChannel(line.Args[0])
		n := conn.GetNick(line.Nick)
		if ch != nil && n != nil {
			ch.DelNick(n)
		} else {
			conn.error("irc.PART(): buh? PART of channel %s by nick %s", line.Args[0], line.Nick)
		}
	})

	// Handle KICKs from channels to maintain state
	conn.AddHandler("KICK", func(conn *Conn, line *Line) {
		// XXX: this won't handle autorejoining channels on KICK
		// it's trivial to do this in a seperate handler...
		ch := conn.GetChannel(line.Args[0])
		n := conn.GetNick(line.Args[1])
		if ch != nil && n != nil {
			ch.DelNick(n)
		} else {
			conn.error("irc.KICK(): buh? KICK from channel %s of nick %s", line.Args[0], line.Args[1])
		}
	})

	// Handle other people's QUITs
	conn.AddHandler("QUIT", func(conn *Conn, line *Line) {
		if n := conn.GetNick(line.Nick); n != nil {
			n.Delete()
		} else {
			conn.error("irc.QUIT(): buh? QUIT from unknown nick %s", line.Nick)
		}
	})

	// Handle MODE changes for channels we know about (and our nick personally)
	// this is moderately ugly. suggestions for improvement welcome
	conn.AddHandler("MODE", func(conn *Conn, line *Line) {
		// channel modes first
		if ch := conn.GetChannel(line.Args[0]); ch != nil {
			modeargs := line.Args[2:len(line.Args)]
			var modeop bool // true => add mode, false => remove mode
			var modestr string
			for i := 0; i < len(line.Args[1]); i++ {
				switch m := line.Args[1][i]; m {
				case '+':
					modeop = true
					modestr = string(m)
				case '-':
					modeop = false
					modestr = string(m)
				case 'i':
					ch.Modes.InviteOnly = modeop
				case 'm':
					ch.Modes.Moderated = modeop
				case 'n':
					ch.Modes.NoExternalMsg = modeop
				case 'p':
					ch.Modes.Private = modeop
				case 's':
					ch.Modes.Secret = modeop
				case 't':
					ch.Modes.ProtectedTopic = modeop
				case 'z':
					ch.Modes.SSLOnly = modeop
				case 'O':
					ch.Modes.OperOnly = modeop
				case 'k':
					if len(modeargs) != 0 {
						ch.Modes.Key, modeargs = modeargs[0], modeargs[1:len(modeargs)]
					} else {
						conn.error("irc.MODE(): buh? not enough arguments to process MODE %s %s%s", ch.Name, modestr, m)
					}
				case 'l':
					if len(modeargs) != 0 {
						ch.Modes.Limit, _ = strconv.Atoi(modeargs[0])
						modeargs = modeargs[1:len(modeargs)]
					} else {
						conn.error("irc.MODE(): buh? not enough arguments to process MODE %s %s%s", ch.Name, modestr, m)
					}
				case 'q', 'a', 'o', 'h', 'v':
					if len(modeargs) != 0 {
						n := conn.GetNick(modeargs[0])
						if p, ok := ch.Nicks[n]; ok && n != nil {
							switch m {
							case 'q':
								p.Owner = modeop
							case 'a':
								p.Admin = modeop
							case 'o':
								p.Op = modeop
							case 'h':
								p.HalfOp = modeop
							case 'v':
								p.Voice = modeop
							}
							modeargs = modeargs[1:len(modeargs)]
						} else {
							conn.error("irc.MODE(): MODE %s %s%s %s: buh? state tracking failure.", ch.Name, modestr, m, modeargs[0])
						}
					} else {
						conn.error("irc.MODE(): buh? not enough arguments to process MODE %s %s%s", ch.Name, modestr, m)
					}
				}
			}
		} else if n := conn.GetNick(line.Args[0]); n != nil {
			// nick mode change, should be us
			if n != conn.Me {
				conn.error("irc.MODE(): buh? recieved MODE %s for (non-me) nick %s", line.Text, n.Nick)
				return
			}
			var modeop bool // true => add mode, false => remove mode
			for i := 0; i < len(line.Text); i++ {
				switch m := line.Text[i]; m {
				case '+':
					modeop = true
				case '-':
					modeop = false
				case 'i':
					n.Modes.Invisible = modeop
				case 'o':
					n.Modes.Oper = modeop
				case 'w':
					n.Modes.WallOps = modeop
				case 'x':
					n.Modes.HiddenHost = modeop
				case 'z':
					n.Modes.SSL = modeop
				}
			}
		} else {
			if line.Text != "" {
				conn.error("irc.MODE(): buh? not sure what to do with nick MODE %s %s", line.Args[0], line.Text)
			} else {
				conn.error("irc.MODE(): buh? not sure what to do with chan MODE %s", strings.Join(line.Args, " "))
			}
		}
	})

	// Handle TOPIC changes for channels
	conn.AddHandler("TOPIC", func(conn *Conn, line *Line) {
		if ch := conn.GetChannel(line.Args[0]); ch != nil {
			ch.Topic = line.Text
		} else {
			conn.error("irc.TOPIC(): buh? topic change on unknown channel %s", line.Args[0])
		}
	})

	// Handle 311 whois reply
	conn.AddHandler("311", func(conn *Conn, line *Line) {
		if n := conn.GetNick(line.Args[1]); n != nil {
			n.Ident = line.Args[2]
			n.Host = line.Args[3]
			n.Name = line.Text
		} else {
			conn.error("irc.311(): buh? received WHOIS info for unknown nick %s", line.Args[1])
		}
	})

	// Handle 324 mode reply
	conn.AddHandler("324", func(conn *Conn, line *Line) {
		// XXX: copypasta from MODE, needs tidying.
		if ch := conn.GetChannel(line.Args[1]); ch != nil {
			modeargs := line.Args[3:len(line.Args)]
			var modeop bool // true => add mode, false => remove mode
			var modestr string
			for i := 0; i < len(line.Args[2]); i++ {
				switch m := line.Args[2][i]; m {
				case '+':
					modeop = true
					modestr = string(m)
				case '-':
					modeop = false
					modestr = string(m)
				case 'i':
					ch.Modes.InviteOnly = modeop
				case 'm':
					ch.Modes.Moderated = modeop
				case 'n':
					ch.Modes.NoExternalMsg = modeop
				case 'p':
					ch.Modes.Private = modeop
				case 's':
					ch.Modes.Secret = modeop
				case 't':
					ch.Modes.ProtectedTopic = modeop
				case 'z':
					ch.Modes.SSLOnly = modeop
				case 'O':
					ch.Modes.OperOnly = modeop
				case 'k':
					if len(modeargs) != 0 {
						ch.Modes.Key, modeargs = modeargs[0], modeargs[1:len(modeargs)]
					} else {
						conn.error("irc.324(): buh? not enough arguments to process MODE %s %s%s", ch.Name, modestr, m)
					}
				case 'l':
					if len(modeargs) != 0 {
						ch.Modes.Limit, _ = strconv.Atoi(modeargs[0])
						modeargs = modeargs[1:len(modeargs)]
					} else {
						conn.error("irc.324(): buh? not enough arguments to process MODE %s %s%s", ch.Name, modestr, m)
					}
				}
			}
		} else {
			conn.error("irc.324(): buh? received MODE settings for unknown channel %s", line.Args[1])
		}
	})

	// Handle 332 topic reply on join to channel
	conn.AddHandler("332", func(conn *Conn, line *Line) {
		if ch := conn.GetChannel(line.Args[1]); ch != nil {
			ch.Topic = line.Text
		} else {
			conn.error("irc.332(): buh? received TOPIC value for unknown channel %s", line.Args[1])
		}
	})

	// Handle 352 who reply
	conn.AddHandler("352", func(conn *Conn, line *Line) {
		if n := conn.GetNick(line.Args[5]); n != nil {
			n.Ident = line.Args[2]
			n.Host = line.Args[3]
			// XXX: do we care about the actual server the nick is on?
			//      or the hop count to this server?
			// line.Text contains "<hop count> <real name>"
			a := strings.Split(line.Text, " ", 2)
			n.Name = a[1]
			if idx := strings.Index(line.Args[6], "*"); idx != -1 {
				n.Modes.Oper = true
			}
			if idx := strings.Index(line.Args[6], "H"); idx != -1 {
				n.Modes.Invisible = true
			}
		} else {
			conn.error("irc.352(): buh? got WHO reply for unknown nick %s", line.Args[5])
		}
	})

	// Handle 353 names reply
	conn.AddHandler("353", func(conn *Conn, line *Line) {
		if ch := conn.GetChannel(line.Args[2]); ch != nil {
			nicks := strings.Split(line.Text, " ", -1)
			for _, nick := range nicks {
				// UnrealIRCd's coders are lazy and leave a trailing space
				if nick == "" {
					continue
				}
				switch c := nick[0]; c {
				case '~', '&', '@', '%', '+':
					nick = nick[1:len(nick)]
					fallthrough
				default:
					n := conn.GetNick(nick)
					if n == nil {
						// we don't know this nick yet!
						n = conn.NewNick(nick, "", "", "")
					}
					if n != conn.Me {
						// we will be in the names list, but should also be in
						// the channel's nick list from the JOIN handler above
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
			conn.error("irc.353(): buh? received NAMES list for unknown channel %s", line.Args[2])
		}
	})

	// Handle 671 whois reply (nick connected via SSL)
	conn.AddHandler("671", func(conn *Conn, line *Line) {
		if n := conn.GetNick(line.Args[1]); n != nil {
			n.Modes.SSL = true
		} else {
			conn.error("irc.671(): buh? received WHOIS SSL info for unknown nick %s", line.Args[1])
		}
	})
}
