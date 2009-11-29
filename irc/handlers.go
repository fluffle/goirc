package irc

// this file contains the basic set of event handlers
// to manage tracking an irc connection etc.

import (
	"fmt";
	"strings";
)

// Add an event handler for a specific IRC command
func (conn *IRCConn) AddHandler(name string, f func (*IRCConn, *IRCLine)) {
	n := strings.ToUpper(name);
	if e, ok := conn.events[n]; ok {
		if len(e) == cap(e) {
			// crap, we're full. expand e by another 10 handler slots
			ne := make([]func (*IRCConn, *IRCLine), len(e), len(e)+10);
			for i := 0; i<len(e); i++ {
				ne[i] = e[i];
			}
			e = ne;
		}
		e = e[0:len(e)+1];
		e[len(e)-1] = f;
	} else {
		e := make([]func (*IRCConn, *IRCLine), 1, 10);
		e[0] = f;
		conn.events[n] = e;
	}
}

// loops through all event handlers for line.Cmd, running each in a goroutine
func (conn *IRCConn) dispatchEvent(line *IRCLine) {
	// So, I think CTCP and (in particular) CTCP ACTION are better handled as
	// separate events as opposed to forcing people to have gargantuan PRIVMSG
	// handlers to cope with the possibilities.
	if line.Cmd == "PRIVMSG" && len(line.Text) > 2
		&& line.Text[0] == '\001' && line.Text[len(line.Text)-1] == '\001' {
		// WOO, it's a CTCP message
		t := strings.Split(line.Text[1:len(line.Text)-1], " ", 2);
		if c := strings.ToUpper(t[0]); c == "ACTION" {
			// make a CTCP ACTION it's own event a-la PRIVMSG
			line.Cmd = c;
		} else {
			// otherwise, dispatch a generic CTCP event that
			// contains the type of CTCP in line.Args[0]
			line.Cmd = "CTCP";
			a := make([]string, len(line.Args)+1);
			a[0] = c;
			for i:=0; i<len(line.Args); i++ {
				a[i+1] = line.Args[i];
			}
			line.Args = a;
		}
		if len(t) > 1 {
			// for some CTCP messages this could make more sense 
			// in line.Args[], but meh. MEH, I say.
			line.Text = t[1];
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
func (conn *IRCConn) setupEvents() {
	// Basic ping/pong handler
	conn.AddHandler("PING",	func(conn *IRCConn, line *IRCLine) {
		conn.send(fmt.Sprintf("PONG :%s", line.Text));
	});

	// Handler to trigger a "CONNECTED" event on receipt of numeric 001
	conn.AddHandler("001", func(conn *IRCConn, line *IRCLine) {
		l := new(IRCLine);
		l.Cmd = "CONNECTED";
		conn.dispatchEvent(l);
	});

	// Handler to deal with "433 :Nickname already in use" on connection
	conn.AddHandler("433", func(conn *IRCConn, line *IRCLine) {
		conn.Nick(conn.Me + "_");
	});
}

