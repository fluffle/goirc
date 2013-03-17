package client

import "strings"

const (
	REGISTER     = "REGISTER"
	CONNECTED    = "CONNECTED"
	DISCONNECTED = "DISCONNECTED"
	ACTION       = "ACTION"
	AWAY         = "AWAY"
	CTCP         = "CTCP"
	CTCPREPLY    = "CTCPREPLY"
	INVITE       = "INVITE"
	JOIN         = "JOIN"
	KICK         = "KICK"
	MODE         = "MODE"
	NICK         = "NICK"
	NOTICE       = "NOTICE"
	OPER         = "OPER"
	PART         = "PART"
	PASS         = "PASS"
	PING         = "PING"
	PONG         = "PONG"
	PRIVMSG      = "PRIVMSG"
	QUIT         = "QUIT"
	TOPIC        = "TOPIC"
	USER         = "USER"
	VERSION      = "VERSION"
	VHOST        = "VHOST"
	WHO          = "WHO"
	WHOIS        = "WHOIS"
)

// cutNewLines() pares down a string to the part before the first "\r" or "\n"
func cutNewLines(s string) string {
	r := strings.SplitN(s, "\r", 2)
	r = strings.SplitN(r[0], "\n", 2)
	return r[0]
}

// indexFragment looks for the last sentence split-point (defined as one of
// the punctuation characters .:;,!?"' followed by a space) in the string s
// and returns the index in the string after that split-point. If no split-
// point is found it returns the index after the last space in s, or -1.
func indexFragment(s string) int {
	max := -1
	for _, sep := range []string{". ", ": ", "; ", ", ", "! ", "? ", "\" ", "' "} {
		if idx := strings.LastIndex(s, sep); idx > max {
			max = idx
		}
	}
	if max > 0 {
		return max + 2
	}
	if idx := strings.LastIndex(s, " "); idx > 0 {
		return idx + 1
	}
	return -1
}

// splitMessage splits a message > splitLen chars at:
//   1. the end of the last sentence fragment before splitLen
//   2. the end of the last word before splitLen
//   3. splitLen itself
func splitMessage(msg string, splitLen int) (msgs []string) {
	// This is quite short ;-)
	if splitLen < 10 {
		splitLen = 10
	}
	for len(msg) > splitLen {
		idx := indexFragment(msg[:splitLen])
		if idx < 0 {
			idx = splitLen
		}
		msgs = append(msgs, msg[:idx] + "...")
		msg = msg[idx:]
	}
	return append(msgs, msg)
}

// Raw() sends a raw line to the server, should really only be used for
// debugging purposes but may well come in handy.
func (conn *Conn) Raw(rawline string) {
	// Avoid command injection by enforcing one command per line.
	conn.out <- cutNewLines(rawline)
}

// Pass() sends a PASS command to the server
func (conn *Conn) Pass(password string) { conn.Raw(PASS + " " + password) }

// Nick() sends a NICK command to the server
func (conn *Conn) Nick(nick string) { conn.Raw(NICK + " " + nick) }

// User() sends a USER command to the server
func (conn *Conn) User(ident, name string) {
	conn.Raw(USER + " " + ident + " 12 * :" + name)
}

// Join() sends a JOIN command to the server
func (conn *Conn) Join(channel string) { conn.Raw(JOIN + " " + channel) }

// Part() sends a PART command to the server with an optional part message
func (conn *Conn) Part(channel string, message ...string) {
	msg := strings.Join(message, " ")
	if msg != "" {
		msg = " :" + msg
	}
	conn.Raw(PART + " " + channel + msg)
}

// Kick() sends a KICK command to remove a nick from a channel
func (conn *Conn) Kick(channel, nick string, message ...string) {
	msg := strings.Join(message, " ")
	if msg != "" {
		msg = " :" + msg
	}
	conn.Raw(KICK + " " + channel + " " + nick + msg)
}

// Quit() sends a QUIT command to the server with an optional quit message
func (conn *Conn) Quit(message ...string) {
	msg := strings.Join(message, " ")
	if msg == "" {
		msg = conn.cfg.QuitMessage
	}
	conn.Raw(QUIT + " :" + msg)
}

// Whois() sends a WHOIS command to the server
func (conn *Conn) Whois(nick string) { conn.Raw(WHOIS + " " + nick) }

//Who() sends a WHO command to the server
func (conn *Conn) Who(nick string) { conn.Raw(WHO + " " + nick) }

// Privmsg() sends a PRIVMSG to the target t
func (conn *Conn) Privmsg(t, msg string) {
	for _, s := range splitMessage(msg, conn.cfg.SplitLen) {
		conn.Raw(PRIVMSG + " " + t + " :" + s)
	}
}

// Notice() sends a NOTICE to the target t
func (conn *Conn) Notice(t, msg string) {
	for _, s := range splitMessage(msg, conn.cfg.SplitLen) {
		conn.Raw(NOTICE + " " + t + " :" + s)
	}
}

// Ctcp() sends a (generic) CTCP message to the target t
// with an optional argument
func (conn *Conn) Ctcp(t, ctcp string, arg ...string) {
	// We need to split again here to ensure
	for _, s := range splitMessage(strings.Join(arg, " "), conn.cfg.SplitLen) {
		if s != "" {
			s = " " + s
		}
		// Using Raw rather than PRIVMSG here to avoid double-split problems.
		conn.Raw(PRIVMSG + " " + t + " :\001" + strings.ToUpper(ctcp) + s + "\001")
	}
}

// CtcpReply() sends a generic CTCP reply to the target t
// with an optional argument
func (conn *Conn) CtcpReply(t, ctcp string, arg ...string) {
	for _, s := range splitMessage(strings.Join(arg, " "), conn.cfg.SplitLen) {
		if s != "" {
			s = " " + s
		}
		// Using Raw rather than NOTICE here to avoid double-split problems.
		conn.Raw(NOTICE + " " + t + " :\001" + strings.ToUpper(ctcp) + s + "\001")
	}
}

// Version() sends a CTCP "VERSION" to the target t
func (conn *Conn) Version(t string) { conn.Ctcp(t, VERSION) }

// Action() sends a CTCP "ACTION" to the target t
func (conn *Conn) Action(t, msg string) { conn.Ctcp(t, ACTION, msg) }

// Topic() sends a TOPIC command to the channel
//   Topic(channel) retrieves the current channel topic (see "332" handler)
//   Topic(channel, topic) sets the topic for the channel
func (conn *Conn) Topic(channel string, topic ...string) {
	t := strings.Join(topic, " ")
	if t != "" {
		t = " :" + t
	}
	conn.Raw(TOPIC + " " + channel + t)
}

// Mode() sends a MODE command to the server. This one can get complicated if
// we try to be too clever, so it's deliberately simple:
//   Mode(t) retrieves the user or channel modes for target t
//   Mode(t, "modestring") sets user or channel modes for target t, where...
//     modestring == e.g. "+o <nick>" or "+ntk <key>" or "-is"
// This means you'll need to do your own mode work. It may be linked in with
// the state tracking and ChanMode/NickMode/ChanPrivs objects later...
func (conn *Conn) Mode(t string, modestring ...string) {
	mode := strings.Join(modestring, " ")
	if mode != "" {
		mode = " " + mode
	}
	conn.Raw(MODE + " " + t + mode)
}

// Away() sends an AWAY command to the server
//   Away() resets away status
//   Away(message) sets away with the given message
func (conn *Conn) Away(message ...string) {
	msg := strings.Join(message, " ")
	if msg != "" {
		msg = " :" + msg
	}
	conn.Raw(AWAY + msg)
}

// Invite() sends an INVITE command to the server
func (conn *Conn) Invite(nick, channel string) {
	conn.Raw(INVITE + " " + nick + " " + channel)
}

// Oper() sends an OPER command to the server
func (conn *Conn) Oper(user, pass string) { conn.Raw(OPER + " " + user + " " + pass) }

// VHost() sends a VHOST command to the server
func (conn *Conn) VHost(user, pass string) { conn.Raw(VHOST + " " + user + " " + pass) }

// Ping() sends a PING command to the server
// A PONG response is to be expected afterwards
func (conn *Conn) Ping(message string) { conn.Raw(PING + " :" + message) }

// Pong() sends a PONG command to the server
func (conn *Conn) Pong(message string) { conn.Raw(PONG + " :" + message) }
