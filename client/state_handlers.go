package client

// this file contains the extra set of event handlers
// to manage tracking state for an IRC connection

import (
	"strings"

	"github.com/fluffle/goirc/logging"
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
		if !conn.Me().Equals(nk) {
			logging.Warn("irc.JOIN(): JOIN to unknown channel %s received "+
				"from (non-me) nick %s", line.Args[0], line.Nick)
			return
		}
		conn.st.NewChannel(line.Args[0])
		// since we don't know much about this channel, ask server for info
		// we get the channel users automatically in 353 and the channel
		// topic in 332 on join, so we just need to get the modes
		conn.Mode(line.Args[0])
		// sending a WHO for the channel is MUCH more efficient than
		// triggering a WHOIS on every nick from the 353 handler
		conn.Who(line.Args[0])
	}
	if nk == nil {
		// this is the first we've seen of this nick
		conn.st.NewNick(line.Nick)
		conn.st.NickInfo(line.Nick, line.Ident, line.Host, "")
		// since we don't know much about this nick, ask server for info
		conn.Who(line.Nick)
	}
	// this takes care of both nick and channel linking \o/
	conn.st.Associate(line.Args[0], line.Nick)
}

// Handle PARTs from channels to maintain state
func (conn *Conn) h_PART(line *Line) {
	conn.st.Dissociate(line.Args[0], line.Nick)
}

// Handle KICKs from channels to maintain state
func (conn *Conn) h_KICK(line *Line) {
	if !line.argslen(1) {
		return
	}
	// XXX: this won't handle autorejoining channels on KICK
	// it's trivial to do this in a seperate handler...
	conn.st.Dissociate(line.Args[0], line.Args[1])
}

// Handle other people's QUITs
func (conn *Conn) h_QUIT(line *Line) {
	conn.st.DelNick(line.Nick)
}

// Handle MODE changes for channels we know about (and our nick personally)
func (conn *Conn) h_MODE(line *Line) {
	if !line.argslen(1) {
		return
	}
	if ch := conn.st.GetChannel(line.Args[0]); ch != nil {
		// channel modes first
		conn.st.ChannelModes(line.Args[0], line.Args[1], line.Args[2:]...)
	} else if nk := conn.st.GetNick(line.Args[0]); nk != nil {
		// nick mode change, should be us
		if !conn.Me().Equals(nk) {
			logging.Warn("irc.MODE(): recieved MODE %s for (non-me) nick %s",
				line.Args[1], line.Args[0])
			return
		}
		conn.st.NickModes(line.Args[0], line.Args[1])
	} else {
		logging.Warn("irc.MODE(): not sure what to do with MODE %s",
			strings.Join(line.Args, " "))
	}
}

// Handle TOPIC changes for channels
func (conn *Conn) h_TOPIC(line *Line) {
	if !line.argslen(1) {
		return
	}
	if ch := conn.st.GetChannel(line.Args[0]); ch != nil {
		conn.st.Topic(line.Args[0], line.Args[1])
	} else {
		logging.Warn("irc.TOPIC(): topic change on unknown channel %s",
			line.Args[0])
	}
}

// Handle 311 whois reply
func (conn *Conn) h_311(line *Line) {
	if !line.argslen(5) {
		return
	}
	if nk := conn.st.GetNick(line.Args[1]); (nk != nil) && !conn.Me().Equals(nk) {
		conn.st.NickInfo(line.Args[1], line.Args[2], line.Args[3], line.Args[5])
	} else {
		logging.Warn("irc.311(): received WHOIS info for unknown nick %s",
			line.Args[1])
	}
}

// Handle 324 mode reply
func (conn *Conn) h_324(line *Line) {
	if !line.argslen(2) {
		return
	}
	if ch := conn.st.GetChannel(line.Args[1]); ch != nil {
		conn.st.ChannelModes(line.Args[1], line.Args[2], line.Args[3:]...)
	} else {
		logging.Warn("irc.324(): received MODE settings for unknown channel %s",
			line.Args[1])
	}
}

// Handle 332 topic reply on join to channel
func (conn *Conn) h_332(line *Line) {
	if !line.argslen(2) {
		return
	}
	if ch := conn.st.GetChannel(line.Args[1]); ch != nil {
		conn.st.Topic(line.Args[1], line.Args[2])
	} else {
		logging.Warn("irc.332(): received TOPIC value for unknown channel %s",
			line.Args[1])
	}
}

// Handle 352 who reply
func (conn *Conn) h_352(line *Line) {
	if !line.argslen(5) {
		return
	}
	nk := conn.st.GetNick(line.Args[5])
	if nk == nil {
		logging.Warn("irc.352(): received WHO reply for unknown nick %s",
			line.Args[5])
		return
	}
	if conn.Me().Equals(nk) {
		return
	}
	// XXX: do we care about the actual server the nick is on?
	//      or the hop count to this server?
	// last arg contains "<hop count> <real name>"
	a := strings.SplitN(line.Args[len(line.Args)-1], " ", 2)
	conn.st.NickInfo(nk.Nick, line.Args[2], line.Args[3], a[1])
	if !line.argslen(6) {
		return
	}
	if idx := strings.Index(line.Args[6], "*"); idx != -1 {
		conn.st.NickModes(nk.Nick, "+o")
	}
	if idx := strings.Index(line.Args[6], "B"); idx != -1 {
		conn.st.NickModes(nk.Nick, "+B")
	}
	if idx := strings.Index(line.Args[6], "H"); idx != -1 {
		conn.st.NickModes(nk.Nick, "+i")
	}
}

// Handle 353 names reply
func (conn *Conn) h_353(line *Line) {
	if !line.argslen(2) {
		return
	}
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
				if conn.st.GetNick(nick) == nil {
					// we don't know this nick yet!
					conn.st.NewNick(nick)
				}
				if _, ok := conn.st.IsOn(ch.Name, nick); !ok {
					// This nick isn't associated with this channel yet!
					conn.st.Associate(ch.Name, nick)
				}
				switch c {
				case '~':
					conn.st.ChannelModes(ch.Name, "+q", nick)
				case '&':
					conn.st.ChannelModes(ch.Name, "+a", nick)
				case '@':
					conn.st.ChannelModes(ch.Name, "+o", nick)
				case '%':
					conn.st.ChannelModes(ch.Name, "+h", nick)
				case '+':
					conn.st.ChannelModes(ch.Name, "+v", nick)
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
	if !line.argslen(1) {
		return
	}
	if nk := conn.st.GetNick(line.Args[1]); nk != nil {
		conn.st.NickModes(nk.Nick, "+z")
	} else {
		logging.Warn("irc.671(): received WHOIS SSL info for unknown nick %s",
			line.Args[1])
	}
}
