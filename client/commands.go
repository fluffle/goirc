package client

import (
	"fmt"
	"strings"
)

const (
	REGISTER     = "REGISTER"
	CONNECTED    = "CONNECTED"
	DISCONNECTED = "DISCONNECTED"
	ACTION       = "ACTION"
	AWAY         = "AWAY"
	CAP          = "CAP"
	CTCP         = "CTCP"
	CTCPREPLY    = "CTCPREPLY"
	ERROR        = "ERROR"
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
	defaultSplit = 450
)

// cutNewLines() pares down a string to the part before the first "\r" or "\n".
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
	if splitLen < 13 {
		splitLen = defaultSplit
	}
	for len(msg) > splitLen {
		idx := indexFragment(msg[:splitLen-3])
		if idx < 0 {
			idx = splitLen - 3
		}
		msgs = append(msgs, msg[:idx]+"...")
		msg = msg[idx:]
	}
	return append(msgs, msg)
}

// Raw sends a raw line to the server, should really only be used for
// debugging purposes but may well come in handy.
func (conn *Conn) Raw(rawline string) {
	// Avoid command injection by enforcing one command per line.
	conn.out <- cutNewLines(rawline)
}

// Pass sends a PASS command to the server.
//     PASS password
func (conn *Conn) Pass(password string) { conn.Raw(PASS + " " + password) }

// Nick sends a NICK command to the server.
//     NICK nick
func (conn *Conn) Nick(nick string) { conn.Raw(NICK + " " + nick) }

// User sends a USER command to the server.
//     USER ident 12 * :name
func (conn *Conn) User(ident, name string) {
	conn.Raw(USER + " " + ident + " 12 * :" + name)
}

// Join sends a JOIN command to the server with an optional key.
//     JOIN channel [key]
func (conn *Conn) Join(channel string, key ...string) {
	k := ""
	if len(key) > 0 {
		k = " " + key[0]
	}
	conn.Raw(JOIN + " " + channel + k)
}

// Part sends a PART command to the server with an optional part message.
//     PART channel [:message]
func (conn *Conn) Part(channel string, message ...string) {
	msg := strings.Join(message, " ")
	if msg != "" {
		msg = " :" + msg
	}
	conn.Raw(PART + " " + channel + msg)
}

// Kick sends a KICK command to remove a nick from a channel.
//     KICK channel nick [:message]
func (conn *Conn) Kick(channel, nick string, message ...string) {
	msg := strings.Join(message, " ")
	if msg != "" {
		msg = " :" + msg
	}
	conn.Raw(KICK + " " + channel + " " + nick + msg)
}

// Quit sends a QUIT command to the server with an optional quit message.
//     QUIT [:message]
func (conn *Conn) Quit(message ...string) {
	msg := strings.Join(message, " ")
	if msg == "" {
		msg = conn.cfg.QuitMessage
	}
	conn.Raw(QUIT + " :" + msg)
}

// Whois sends a WHOIS command to the server.
//     WHOIS nick
func (conn *Conn) Whois(nick string) { conn.Raw(WHOIS + " " + nick) }

// Who sends a WHO command to the server.
//     WHO nick
func (conn *Conn) Who(nick string) { conn.Raw(WHO + " " + nick) }

// Privmsg sends a PRIVMSG to the target nick or channel t.
// If msg is longer than Config.SplitLen characters, multiple PRIVMSGs
// will be sent to the target containing sequential parts of msg.
// PRIVMSG t :msg
func (conn *Conn) Privmsg(t, msg string) {
	prefix := PRIVMSG + " " + t + " :"
	for _, s := range splitMessage(msg, conn.cfg.SplitLen) {
		conn.Raw(prefix + s)
	}
}

// Privmsgln is the variadic version of Privmsg that formats the message
// that is sent to the target nick or channel t using the
// fmt.Sprintln function.
// Note: Privmsgln doesn't add the '\n' character at the end of the message.
func (conn *Conn) Privmsgln(t string, a ...interface{}) {
	msg := fmt.Sprintln(a...)
	// trimming the new-line character added by the fmt.Sprintln function,
	// since it's irrelevant.
	msg = msg[:len(msg)-1]
	conn.Privmsg(t, msg)
}

// Privmsgf is the variadic version of Privmsg that formats the message
// that is sent to the target nick or channel t using the
// fmt.Sprintf function.
func (conn *Conn) Privmsgf(t, format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	conn.Privmsg(t, msg)
}

// Notice sends a NOTICE to the target nick or channel t.
// If msg is longer than Config.SplitLen characters, multiple NOTICEs
// will be sent to the target containing sequential parts of msg.
//     NOTICE t :msg
func (conn *Conn) Notice(t, msg string) {
	for _, s := range splitMessage(msg, conn.cfg.SplitLen) {
		conn.Raw(NOTICE + " " + t + " :" + s)
	}
}

// Ctcp sends a (generic) CTCP message to the target nick
// or channel t, with an optional argument.
//     PRIVMSG t :\001CTCP arg\001
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

// CtcpReply sends a (generic) CTCP reply to the target nick
// or channel t, with an optional argument.
//     NOTICE t :\001CTCP arg\001
func (conn *Conn) CtcpReply(t, ctcp string, arg ...string) {
	for _, s := range splitMessage(strings.Join(arg, " "), conn.cfg.SplitLen) {
		if s != "" {
			s = " " + s
		}
		// Using Raw rather than NOTICE here to avoid double-split problems.
		conn.Raw(NOTICE + " " + t + " :\001" + strings.ToUpper(ctcp) + s + "\001")
	}
}

// Version sends a CTCP "VERSION" to the target nick or channel t.
func (conn *Conn) Version(t string) { conn.Ctcp(t, VERSION) }

// Action sends a CTCP "ACTION" to the target nick or channel t.
func (conn *Conn) Action(t, msg string) { conn.Ctcp(t, ACTION, msg) }

// Topic() sends a TOPIC command for a channel.
// If no topic is provided this requests that a 332 response is sent by the
// server for that channel, which can then be handled to retrieve the current
// channel topic. If a topic is provided the channel's topic will be set.
//     TOPIC channel
//     TOPIC channel :topic
func (conn *Conn) Topic(channel string, topic ...string) {
	t := strings.Join(topic, " ")
	if t != "" {
		t = " :" + t
	}
	conn.Raw(TOPIC + " " + channel + t)
}

// Mode sends a MODE command for a target nick or channel t.
// If no mode strings are provided this requests that a 324 response is sent
// by the server for the target. Otherwise the mode strings are concatenated
// with spaces and sent to the server. This allows e.g.
//     conn.Mode("#channel", "+nsk", "mykey")
//
//     MODE t
//     MODE t modestring
func (conn *Conn) Mode(t string, modestring ...string) {
	mode := strings.Join(modestring, " ")
	if mode != "" {
		mode = " " + mode
	}
	conn.Raw(MODE + " " + t + mode)
}

// Away sends an AWAY command to the server.
// If a message is provided it sets the client's away status with that message,
// otherwise it resets the client's away status.
//     AWAY
//     AWAY :message
func (conn *Conn) Away(message ...string) {
	msg := strings.Join(message, " ")
	if msg != "" {
		msg = " :" + msg
	}
	conn.Raw(AWAY + msg)
}

// Invite sends an INVITE command to the server.
//     INVITE nick channel
func (conn *Conn) Invite(nick, channel string) {
	conn.Raw(INVITE + " " + nick + " " + channel)
}

// Oper sends an OPER command to the server.
//     OPER user pass
func (conn *Conn) Oper(user, pass string) { conn.Raw(OPER + " " + user + " " + pass) }

// VHost sends a VHOST command to the server.
//     VHOST user pass
func (conn *Conn) VHost(user, pass string) { conn.Raw(VHOST + " " + user + " " + pass) }

// Ping sends a PING command to the server, which should PONG.
//     PING :message
func (conn *Conn) Ping(message string) { conn.Raw(PING + " :" + message) }

// Pong sends a PONG command to the server.
//     PONG :message
func (conn *Conn) Pong(message string) { conn.Raw(PONG + " :" + message) }

// Cap sends a CAP command to the server.
//     CAP subcommand
//     CAP subcommand :message
func (conn *Conn) Cap(subcommmand string, capabilities ...string) {
	if len(capabilities) == 0 {
		conn.Raw(CAP + " " + subcommmand)
	} else {
		conn.Raw(CAP + " " + subcommmand + " :" + strings.Join(capabilities, " "))
	}
}
