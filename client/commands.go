package client

import "strings"

// this file contains the various commands you can
// send to the server using an Conn connection

// This could be a lot less ugly with the ability to manipulate
// the symbol table and add methods/functions on the fly
// [ CMD, FMT, FMTARGS ] etc.

// safe removes CR/LF to avoid IRC command injection.
func safe(s string) string {
	s = strings.Replace(s, "\r", "", -1)
	s = strings.Replace(s, "\n", "", -1)
	return s
}

// Raw() sends a raw line to the server, should really only be used for
// debugging purposes but may well come in handy.
func (conn *Conn) Raw(rawline string) { conn.out <- rawline }

// Pass() sends a PASS command to the server
func (conn *Conn) Pass(password string) { conn.out <- safe("PASS " + password) }

// Nick() sends a NICK command to the server
func (conn *Conn) Nick(nick string) { conn.out <- safe("NICK " + nick) }

// User() sends a USER command to the server
func (conn *Conn) User(ident, name string) {
	conn.out <- safe("USER " + ident + " 12 * :" + name)
}

// Join() sends a JOIN command to the server
func (conn *Conn) Join(channel string) { conn.out <- safe("JOIN " + channel) }

// Part() sends a PART command to the server with an optional part message
func (conn *Conn) Part(channel string, message ...string) {
	msg := strings.Join(message, " ")
	if msg != "" {
		msg = " :" + msg
	}
	conn.out <- safe("PART " + channel + msg)
}

// Kick() sends a KICK command to remove a nick from a channel
func (conn *Conn) Kick(channel, nick string, message ...string) {
	msg := strings.Join(message, " ")
	if msg != "" {
		msg = " :" + msg
	}
	conn.out <- safe("KICK " + channel + " " + nick + msg)
}

// Quit() sends a QUIT command to the server with an optional quit message
func (conn *Conn) Quit(message ...string) {
	msg := strings.Join(message, " ")
	if msg == "" {
		msg = conn.cfg.QuitMessage
	}
	conn.out <- safe("QUIT :" + msg)
}

// Whois() sends a WHOIS command to the server
func (conn *Conn) Whois(nick string) { conn.out <- safe("WHOIS " + nick) }

//Who() sends a WHO command to the server
func (conn *Conn) Who(nick string) { conn.out <- safe("WHO " + nick) }

// Privmsg() sends a PRIVMSG to the target t
func (conn *Conn) Privmsg(t, msg string) { conn.out <- safe("PRIVMSG " + t + " :" + msg) }

// Notice() sends a NOTICE to the target t
func (conn *Conn) Notice(t, msg string) { conn.out <- safe("NOTICE " + t + " :" + msg) }

// Ctcp() sends a (generic) CTCP message to the target t
// with an optional argument
func (conn *Conn) Ctcp(t, ctcp string, arg ...string) {
	msg := strings.Join(arg, " ")
	if msg != "" {
		msg = " " + msg
	}
	conn.Privmsg(t, "\001"+strings.ToUpper(ctcp)+msg+"\001")
}

// CtcpReply() sends a generic CTCP reply to the target t
// with an optional argument
func (conn *Conn) CtcpReply(t, ctcp string, arg ...string) {
	msg := strings.Join(arg, " ")
	if msg != "" {
		msg = " " + msg
	}
	conn.Notice(t, "\001"+strings.ToUpper(ctcp)+msg+"\001")
}

// Version() sends a CTCP "VERSION" to the target t
func (conn *Conn) Version(t string) { conn.Ctcp(t, "VERSION") }

// Action() sends a CTCP "ACTION" to the target t
func (conn *Conn) Action(t, msg string) { conn.Ctcp(t, "ACTION", msg) }

// Topic() sends a TOPIC command to the channel
//   Topic(channel) retrieves the current channel topic (see "332" handler)
//   Topic(channel, topic) sets the topic for the channel
func (conn *Conn) Topic(channel string, topic ...string) {
	t := strings.Join(topic, " ")
	if t != "" {
		t = " :" + t
	}
	conn.out <- safe("TOPIC " + channel + t)
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
	conn.out <- safe("MODE " + t + mode)
}

// Away() sends an AWAY command to the server
//   Away() resets away status
//   Away(message) sets away with the given message
func (conn *Conn) Away(message ...string) {
	msg := strings.Join(message, " ")
	if msg != "" {
		msg = " :" + msg
	}
	conn.out <- safe("AWAY" + msg)
}

// Invite() sends an INVITE command to the server
func (conn *Conn) Invite(nick, channel string) {
	conn.out <- safe("INVITE " + nick + " " + channel)
}

// Oper() sends an OPER command to the server
func (conn *Conn) Oper(user, pass string) {
	conn.out <- safe("OPER " + user + " " + pass)
}
