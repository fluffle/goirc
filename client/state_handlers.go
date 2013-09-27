package client

// this file contains the extra set of event handlers
// to manage tracking state for an IRC connection

import (
	"github.com/fluffle/goirc/logging"
	"strings"
)

var stHandlers = map[string]HandlerFunc{
	"JOIN":  (*Conn).h_JOIN,
	"KICK":  (*Conn).h_KICK,
	"MODE":  (*Conn).h_MODE,
	"NICK":  (*Conn).h_STNICK,
	"PART":  (*Conn).h_PART,
	"QUIT":  (*Conn).h_QUIT,
	"TOPIC": (*Conn).h_TOPIC,
	"311":   (*Conn).h_311,
	"324":   (*Conn).h_324,
	"332":   (*Conn).h_332,
	"352":   (*Conn).h_352,
	"353":   (*Conn).h_353,
	"671":   (*Conn).h_671,
}

func (conn *Conn) addSTHandlers() {
	for n, h := range stHandlers {
		conn.stRemovers = append(conn.stRemovers, conn.handle(n, h))
	}
}

func (conn *Conn) delSTHandlers() {
	for _, h := range conn.stRemovers {
		h.Remove()
	}
	conn.stRemovers = conn.stRemovers[:0]
}

// Handle NICK messages that need to update the state tracker
func (conn *Conn) h_STNICK(line *Line) {
	// all nicks should be handled the same way, our own included
	conn.st.ReNick(line.Nick, line.Args[0])
}

// Handle JOINs to channels to maintain state
func (conn *Conn) h_JOIN(line *Line) {
	ch := conn.st.GetChannel(line.Args[0])
	nk := conn.st.GetNick(line.Nick)
	if ch == nil {
		// first we've seen of this channel, so should be us joining it
		// NOTE this will also take care of nk == nil && ch == nil
		if nk != conn.cfg.Me {
			logging.Warn("irc.JOIN(): JOIN to unknown channel %s received "+
				"from (non-me) nick %s", line.Args[0], line.Nick)
			return
		}
		ch = conn.st.NewChannel(line.Args[0])
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
		nk = conn.st.NewNick(line.Nick)
		nk.Ident = line.Ident
		nk.Host = line.Host
		// since we don't know much about this nick, ask server for info
		conn.Who(nk.Nick)
	}
	// this takes care of both nick and channel linking \o/
	conn.st.Associate(ch, nk)
}

// Handle PARTs from channels to maintain state
func (conn *Conn) h_PART(line *Line) {
	conn.st.Dissociate(conn.st.GetChannel(line.Args[0]),
		conn.st.GetNick(line.Nick))
}

// Handle KICKs from channels to maintain state
func (conn *Conn) h_KICK(line *Line) {
	// XXX: this won't handle autorejoining channels on KICK
	// it's trivial to do this in a seperate handler...
	conn.st.Dissociate(conn.st.GetChannel(line.Args[0]),
		conn.st.GetNick(line.Args[1]))
}

// Handle other people's QUITs
func (conn *Conn) h_QUIT(line *Line) {
	conn.st.DelNick(line.Nick)
}

// Handle MODE changes for channels we know about (and our nick personally)
func (conn *Conn) h_MODE(line *Line) {
	if ch := conn.st.GetChannel(line.Args[0]); ch != nil {
		// channel modes first
		ch.ParseModes(line.Args[1], line.Args[2:]...)
	} else if nk := conn.st.GetNick(line.Args[0]); nk != nil {
		// nick mode change, should be us
		if nk != conn.cfg.Me {
			logging.Warn("irc.MODE(): recieved MODE %s for (non-me) nick %s",
				line.Args[1], line.Args[0])
			return
		}
		nk.ParseModes(line.Args[1])
	} else {
		logging.Warn("irc.MODE(): not sure what to do with MODE %s",
			strings.Join(line.Args, " "))
	}
}

// Handle TOPIC changes for channels
func (conn *Conn) h_TOPIC(line *Line) {
	if ch := conn.st.GetChannel(line.Args[0]); ch != nil {
		ch.Topic = line.Args[1]
	} else {
		logging.Warn("irc.TOPIC(): topic change on unknown channel %s",
			line.Args[0])
	}
}

// Handle 311 whois reply
func (conn *Conn) h_311(line *Line) {
	if nk := conn.st.GetNick(line.Args[1]); nk != nil && nk != conn.Me() {
		nk.Ident = line.Args[2]
		nk.Host = line.Args[3]
		nk.Name = line.Args[5]
	} else {
		logging.Warn("irc.311(): received WHOIS info for unknown nick %s",
			line.Args[1])
	}
}

// Handle 324 mode reply
func (conn *Conn) h_324(line *Line) {
	if ch := conn.st.GetChannel(line.Args[1]); ch != nil {
		ch.ParseModes(line.Args[2], line.Args[3:]...)
	} else {
		logging.Warn("irc.324(): received MODE settings for unknown channel %s",
			line.Args[1])
	}
}

// Handle 332 topic reply on join to channel
func (conn *Conn) h_332(line *Line) {
	if ch := conn.st.GetChannel(line.Args[1]); ch != nil {
		ch.Topic = line.Args[2]
	} else {
		logging.Warn("irc.332(): received TOPIC value for unknown channel %s",
			line.Args[1])
	}
}

// Handle 352 who reply
func (conn *Conn) h_352(line *Line) {
	nk := conn.st.GetNick(line.Args[5])
	if nk == nil {
		logging.Warn("irc.352(): received WHO reply for unknown nick %s",
			line.Args[5])
		return
	}
	if nk == conn.Me() {
		return
	}
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
}

// Handle 353 names reply
func (conn *Conn) h_353(line *Line) {
	if ch := conn.st.GetChannel(line.Args[2]); ch != nil {
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
				nk := conn.st.GetNick(nick)
				if nk == nil {
					// we don't know this nick yet!
					nk = conn.st.NewNick(nick)
				}
				cp, ok := conn.st.IsOn(ch.Name, nick)
				if !ok {
					// This nick isn't associated with this channel yet!
					cp = conn.st.Associate(ch, nk)
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
		logging.Warn("irc.353(): received NAMES list for unknown channel %s",
			line.Args[2])
	}
}

// Handle 671 whois reply (nick connected via SSL)
func (conn *Conn) h_671(line *Line) {
	if nk := conn.st.GetNick(line.Args[1]); nk != nil {
		nk.Modes.SSL = true
	} else {
		logging.Warn("irc.671(): received WHOIS SSL info for unknown nick %s",
			line.Args[1])
	}
}
