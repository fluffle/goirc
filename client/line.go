package client

import (
	"runtime"
	"strings"
	"time"

	"github.com/fluffle/goirc/logging"
)

var tagsReplacer = strings.NewReplacer("\\:", ";", "\\s", " ", "\\r", "\r", "\\n", "\n")

// We parse an incoming line into this struct. Line.Cmd is used as the trigger
// name for incoming event handlers and is the IRC verb, the first sequence
// of non-whitespace characters after ":nick!user@host", e.g. PRIVMSG.
//   Raw =~ ":nick!user@host cmd args[] :text"
//   Src == "nick!user@host"
//   Cmd == e.g. PRIVMSG, 332
type Line struct {
	Tags                   map[string]string
	Nick, Ident, Host, Src string
	Cmd, Raw               string
	Args                   []string
	Time                   time.Time
}

// Copy returns a deep copy of the Line.
func (l *Line) Copy() *Line {
	nl := *l
	nl.Args = make([]string, len(l.Args))
	copy(nl.Args, l.Args)
	if l.Tags != nil {
		nl.Tags = make(map[string]string)
		for k, v := range l.Tags {
			nl.Tags[k] = v
		}
	}
	return &nl
}

// Text returns the contents of the text portion of a line. This only really
// makes sense for lines with a :text part, but there are a lot of them.
func (line *Line) Text() string {
	if len(line.Args) > 0 {
		return line.Args[len(line.Args)-1]
	}
	return ""
}

// Target returns the contextual target of the line, usually the first Arg
// for the IRC verb. If the line was broadcast from a channel, the target
// will be that channel. If the line was sent directly by a user, the target
// will be that user.
func (line *Line) Target() string {
	// TODO(fluffle): Add 005 CHANTYPES parsing for this?
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

// Public returns true if the line is the result of an IRC user sending
// a message to a channel the client has joined instead of directly
// to the client.
//
// NOTE: This is very permissive, allowing all 4 RFC channel types even if
// your server doesn't technically support them.
func (line *Line) Public() bool {
	switch line.Cmd {
	case PRIVMSG, NOTICE, ACTION:
		switch line.Args[0][0] {
		case '#', '&', '+', '!':
			return true
		}
	case CTCP, CTCPREPLY:
		// CTCP prepends the CTCP verb to line.Args, thus for the message
		//   :nick!user@host PRIVMSG #foo :\001BAR baz\001
		// line.Args contains: []string{"BAR", "#foo", "baz"}
		// TODO(fluffle): Arguably this is broken, and we should have
		// line.Args containing: []string{"#foo", "BAR", "baz"}
		// ... OR change conn.Ctcp()'s argument order to be consistent.
		switch line.Args[1][0] {
		case '#', '&', '+', '!':
			return true
		}
	}
	return false
}

// ParseLine creates a Line from an incoming message from the IRC server.
//
// It contains special casing for CTCP messages, most notably CTCP ACTION.
// All CTCP messages have the \001 bytes stripped from the message and the
// CTCP command separated from any subsequent text. Then, CTCP ACTIONs are
// rewritten such that Line.Cmd == ACTION. Other CTCP messages have Cmd
// set to CTCP or CTCPREPLY, and the CTCP command prepended to line.Args.
//
// ParseLine also parses IRCv3 tags, if received. If a line does not have
// the tags section, Line.Tags will be nil. Tags are optional, and will
// only be included after the correct CAP command.
//
// http://ircv3.net/specs/core/capability-negotiation-3.1.html
// http://ircv3.net/specs/core/message-tags-3.2.html
func ParseLine(s string) *Line {
	line := &Line{Raw: s}

	if s == "" {
		return nil
	}

	if s[0] == '@' {
		var rawTags string
		line.Tags = make(map[string]string)
		if idx := strings.Index(s, " "); idx != -1 {
			rawTags, s = s[1:idx], s[idx+1:]
		} else {
			return nil
		}

		// ; is represented as \: in a tag, so it's safe to split on ;
		for _, tag := range strings.Split(rawTags, ";") {
			if tag == "" {
				continue
			}

			pair := strings.SplitN(tagsReplacer.Replace(tag), "=", 2)
			if len(pair) < 2 {
				line.Tags[tag] = ""
			} else {
				line.Tags[pair[0]] = pair[1]
			}
		}
	}

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

func (line *Line) argslen(minlen int) bool {
	pc, _, _, _ := runtime.Caller(1)
	fn := runtime.FuncForPC(pc)
	if len(line.Args) <= minlen {
		logging.Warn("%s: too few arguments: %s", fn.Name(), strings.Join(line.Args, " "))
		return false
	}
	return true
}
