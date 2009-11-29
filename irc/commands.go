package irc

// this file contains the various commands you can
// send to the server using an IRCConn connection

import (
	"fmt";
	"reflect";
)

// This could be a lot less ugly with the ability to manipulate
// the symbol table and add methods/functions on the fly
// [ CMD, FMT, FMTARGS ] etc. 

// send a PASS command to the server
func (conn *IRCConn) Pass(p string) {
	conn.send(fmt.Sprintf("PASS %s", p));
}

// send a NICK command to the server
func (conn *IRCConn) Nick(n string) {
	conn.send(fmt.Sprintf("NICK %s", n));
}

// send a USER command to the server
func (conn *IRCConn) User(u, n string) {
	conn.send(fmt.Sprintf("USER %s 12 * :%s", u, n));
}

// send a JOIN command to the server
func (conn *IRCConn) Join(c string) {
	conn.send(fmt.Sprintf("JOIN %s", c));
}

// send a PART command to the server
func (conn *IRCConn) Part(c string, a ...) {
	msg := getStringMsg(a);
	if msg != "" {
		msg = " :" + msg
	}
	conn.send(fmt.Sprintf("PART %s%s", c, msg));
}

// send a QUIT command to the server
func (conn *IRCConn) Quit(a ...) {
	msg := getStringMsg(a);
	if msg == "" {
		msg = "GoBye!"
	}
	conn.send(fmt.Sprintf("QUIT :%s", msg));
}

// send a PRIVMSG to the target t
func (conn *IRCConn) Privmsg(t, msg string) {
	conn.send(fmt.Sprintf("PRIVMSG %s :%s", t, msg));
}

// send a NOTICE to the target t
func (conn *IRCConn) Notice(t, msg string) {
	conn.send(fmt.Sprintf("NOTICE %s :%s", t, msg));
}

// send a (generic) CTCP to the target t
func (conn *IRCConn) Ctcp(t, ctcp string, a ...) {
	msg := getStringMsg(a);
	if msg != "" {
		msg = " " + msg
	}
	conn.Privmsg(t, fmt.Sprintf("\001%s%s\001", ctcp, msg));
}

// send a generic CTCP reply to the target t
func (conn *IRCConn) CtcpReply(t, ctcp string, a ...) {
	msg := getStringMsg(a);
	if msg != "" {
		msg = " " + msg
	}
	conn.Notice(t, fmt.Sprintf("\001%s%s\001", ctcp, msg));
}


// send a CTCP "VERSION" to the target t
func (conn *IRCConn) Version(t string) {
	conn.Ctcp(t, "VERSION");
}

// send a CTCP "ACTION" to the target t -- /me does stuff!
func (conn *IRCConn) Action(t, msg string) {
	conn.Ctcp(t, "ACTION", msg);
}

func getStringMsg(a ...) (msg string) {
	// dealing with functions with a variable parameter list is nasteeh :-(
	// the below stolen and munged from fmt/print.go
	if v := reflect.NewValue(a).(*reflect.StructValue); v.NumField() == 1 {
		// XXX: should we check that this looks at least vaguely stringy first?
		msg = fmt.Sprintf("%v", v.Field(1));
	} else {
		msg = ""
	}
	return
}
