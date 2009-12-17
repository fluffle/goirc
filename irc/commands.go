package irc

// this file contains the various commands you can
// send to the server using an Conn connection

import (
	//	"fmt";
	"reflect"
)

// This could be a lot less ugly with the ability to manipulate
// the symbol table and add methods/functions on the fly
// [ CMD, FMT, FMTARGS ] etc.

// send a raw line to the server for debugging etc
func (conn *Conn) Raw(s string) { conn.out <- s }

// send a PASS command to the server
func (conn *Conn) Pass(p string) { conn.out <- "PASS "+p }

// send a NICK command to the server
func (conn *Conn) Nick(n string) { conn.out <- "NICK "+n }

// send a USER command to the server
func (conn *Conn) User(u, n string) { conn.out <- "USER "+u+" 12 * :"+n }

// send a JOIN command to the server
func (conn *Conn) Join(c string) { conn.out <- "JOIN "+c }

// send a PART command to the server
func (conn *Conn) Part(c string, a ...) {
	msg := getStringMsg(a)
	if msg != "" {
		msg = " :" + msg
	}
	conn.out <- "PART "+c+msg
}

// send a QUIT command to the server
func (conn *Conn) Quit(a ...) {
	msg := getStringMsg(a)
	if msg == "" {
		msg = "GoBye!"
	}
	conn.out <- "QUIT :"+msg
}

// send a WHOIS command to the server
func (conn *Conn) Whois(t string) { conn.out <- "WHOIS "+t }

// send a WHO command to the server
func (conn *Conn) Who(t string) { conn.out <- "WHO "+t }

// send a PRIVMSG to the target t
func (conn *Conn) Privmsg(t, msg string) { conn.out <- "PRIVMSG "+t+" :"+msg }

// send a NOTICE to the target t
func (conn *Conn) Notice(t, msg string) { conn.out <- "NOTICE "+t+" :"+msg }

// send a (generic) CTCP to the target t
func (conn *Conn) Ctcp(t, ctcp string, a ...) {
	msg := getStringMsg(a)
	if msg != "" {
		msg = " " + msg
	}
	conn.Privmsg(t, "\001"+ctcp+msg+"\001")
}

// send a generic CTCP reply to the target t
func (conn *Conn) CtcpReply(t, ctcp string, a ...) {
	msg := getStringMsg(a)
	if msg != "" {
		msg = " " + msg
	}
	conn.Notice(t, "\001"+ctcp+msg+"\001")
}

// send a CTCP "VERSION" to the target t
func (conn *Conn) Version(t string) { conn.Ctcp(t, "VERSION") }

// send a CTCP "ACTION" to the target t -- /me does stuff!
func (conn *Conn) Action(t, msg string) { conn.Ctcp(t, "ACTION", msg) }

// send a TOPIC command to the channel c
func (conn *Conn) Topic(c string, a ...) {
	topic := getStringMsg(a)
	if topic != "" {
		topic = " :" + topic
	}
	conn.out <- "TOPIC "+c+topic
}

// send a MODE command (this one gets complicated)
// Mode(t) retrieves the user or channel modes for target t
// Mode(t, "string"
func (conn *Conn) Mode(t string, a ...) {
	mode := getStringMsg(a)
	if mode != "" {
		mode = " " + mode
	}
	conn.out <- "MODE "+t+mode
}

func getStringMsg(a ...) (msg string) {
	// dealing with functions with a variable parameter list is nasteeh :-(
	// the below stolen and munged from fmt/print.go func getString()
	if v := reflect.NewValue(a).(*reflect.StructValue); v.NumField() > 0 {
		if s, ok := v.Field(0).(*reflect.StringValue); ok {
			return s.Get()
		}
		if b, ok := v.Interface().([]byte); ok {
			return string(b)
		}
	}
	return ""
}
