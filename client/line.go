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
	Time                   *time.Time
}

func parseLine(s string) *Line {
	line := &Line{Raw: s}
	if s[0] == ':' {
		// remove a source and parse it
		if idx := strings.Index(s, " "); idx != -1 {
			line.Src, s = s[1:idx], s[idx+1:len(s)]
		} else {
			// pretty sure we shouldn't get here ...
			line.Src = s[1:len(s)]
			return line
		}

		// src can be the hostname of the irc server or a nick!user@host
		line.Host = line.Src
		nidx, uidx := strings.Index(line.Src, "!"), strings.Index(line.Src, "@")
		if uidx != -1 && nidx != -1 {
			line.Nick = line.Src[0:nidx]
			line.Ident = line.Src[nidx+1 : uidx]
			line.Host = line.Src[uidx+1 : len(line.Src)]
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
		line.Args = args[1:len(args)]
	}
	return line
}
