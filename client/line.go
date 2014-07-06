package client

import (
	"strings"
	"time"
)

// We parse an incoming line into this struct. Line.Cmd is used as the trigger
// name for incoming event handlers, see *Conn.recv() for details.
//   Raw =~ ":nick!user@host cmd args[] :text"
//   Src == "nick!user@host"
//   Cmd == e.g. PRIVMSG, 332
type Line struct {
	Nick, Ident, Host, Src string
	Cmd, Raw               string
	Args                   []string
	Time                   time.Time
}

// Copy() returns a deep copy of the Line.
func (l *Line) Copy() *Line {
	nl := *l
	nl.Args = make([]string, len(l.Args))
	copy(nl.Args, l.Args)
	return &nl
}

// Return the contents of the text portion of a line.  This only really
// makes sense for lines with a :text part, but there are a lot of them.
func (line *Line) Text() string {
	if len(line.Args) > 0 {
		return line.Args[len(line.Args)-1]
	}
	return ""
}

// Return the target of the line, usually the first Arg for the IRC verb.
// If the line was broadcast from a channel, the target will be that channel.
// If the line was broadcast by a user, the target will be that user.
// TODO(fluffle): Add 005 CHANTYPES parsing for this?
func (line *Line) Target() string {
	switch line.Cmd {
	case PRIVMSG, NOTICE, ACTION:
		if !line.Public() {
			return line.Nick
		}
	case CTCP, CTCPREPLY:
		if !line.Public() {
			return line.Nick
		}
		return line.Args[1]
	}
	if len(line.Args) > 0 {
		return line.Args[0]
	}
	return ""
}

// NOTE: Makes the assumption that all channels start with #.
func (line *Line) Public() bool {
	switch line.Cmd {
	case PRIVMSG, NOTICE, ACTION:
		if strings.HasPrefix(line.Args[0], "#") {
			return true
		}
	case CTCP, CTCPREPLY:
		// CTCP prepends the CTCP verb to line.Args, thus for the message
		//   :nick!user@host PRIVMSG #foo :\001BAR baz\001
		// line.Args contains: []string{"BAR", "#foo", "baz"}
		// TODO(fluffle): Arguably this is broken, and we should have
		// line.Args containing: []string{"#foo", "BAR", "baz"}
		// ... OR change conn.Ctcp()'s argument order to be consistent.
		if strings.HasPrefix(line.Args[1], "#") {
			return true
		}
	}
	return false
}


// ParseLine() creates a Line from an incoming message from the IRC server.
func ParseLine(s string) *Line {
	line := &Line{Raw: s}
	if s[0] == ':' {
		// remove a source and parse it
		if idx := strings.Index(s, " "); idx != -1 {
			line.Src, s = s[1:idx], s[idx+1:]
		} else {
			// pretty sure we shouldn't get here ...
			return nil
		}

		// src can be the hostname of the irc server or a nick!user@host
		line.Host = line.Src
		nidx, uidx := strings.Index(line.Src, "!"), strings.Index(line.Src, "@")
		if uidx != -1 && nidx != -1 {
			line.Nick = line.Src[:nidx]
			line.Ident = line.Src[nidx+1 : uidx]
			line.Host = line.Src[uidx+1:]
		}
	}

	// now we're here, we've parsed a :nick!user@host or :server off
	// s should contain "cmd args[] :text"
	args := strings.SplitN(s, " :", 2)
	if len(args) > 1 {
		args = append(strings.Fields(args[0]), args[1])
	} else {
		args = strings.Fields(args[0])
	}
	line.Cmd = strings.ToUpper(args[0])
	if len(args) > 1 {
		line.Args = args[1:]
	}

	// So, I think CTCP and (in particular) CTCP ACTION are better handled as
	// separate events as opposed to forcing people to have gargantuan
	// handlers to cope with the possibilities.
	if (line.Cmd == PRIVMSG || line.Cmd == NOTICE) &&
		len(line.Args[1]) > 2 &&
		strings.HasPrefix(line.Args[1], "\001") &&
		strings.HasSuffix(line.Args[1], "\001") {
		// WOO, it's a CTCP message
		t := strings.SplitN(strings.Trim(line.Args[1], "\001"), " ", 2)
		if len(t) > 1 {
			// Replace the line with the unwrapped CTCP
			line.Args[1] = t[1]
		}
		if c := strings.ToUpper(t[0]); c == ACTION && line.Cmd == PRIVMSG {
			// make a CTCP ACTION it's own event a-la PRIVMSG
			line.Cmd = c
		} else {
			// otherwise, dispatch a generic CTCP/CTCPREPLY event that
			// contains the type of CTCP in line.Args[0]
			if line.Cmd == PRIVMSG {
				line.Cmd = CTCP
			} else {
				line.Cmd = CTCPREPLY
			}
			line.Args = append([]string{c}, line.Args...)
		}
	}
	return line
}
