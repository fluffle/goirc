package state

import (
	"github.com/fluffle/goirc/logging"
	"reflect"
)

// A struct representing an IRC nick
type Nick struct {
	Nick, Ident, Host, Name string
	Modes                   *NickMode
	chans                   map[*Channel]*ChanPrivs
	me                      bool
	st                      StateTracker
}

// A struct representing the modes of an IRC Nick (User Modes)
// (again, only the ones we care about)
//
// This is only really useful for me, as we can't see other people's modes
// without IRC operator privileges (and even then only on some IRCd's).
type NickMode struct {
	// MODE +i, +o, +w, +x, +z
	Invisible, Oper, WallOps, HiddenHost, SSL bool
}

// Map *irc.NickMode fields to IRC mode characters and vice versa
var StringToNickMode = map[string]string{}
var NickModeToString = map[string]string{
	"Invisible":  "i",
	"Oper":       "o",
	"WallOps":    "w",
	"HiddenHost": "x",
	"SSL":        "z",
}

func init() {
	for k, v := range NickModeToString {
		StringToNickMode[v] = k
	}
}

/******************************************************************************\
 * Nick methods for state management
\******************************************************************************/

func (n *Nick) initialise() {
	n.Modes = new(NickMode)
	n.chans = make(map[*Channel]*ChanPrivs)
}

// Associates a Channel with a Nick using a shared ChanPrivs
//
// Very slightly different to Channel.AddNick() in that it tests for a
// pre-existing association within the Nick object rather than the
// Channel object before associating the two.
func (n *Nick) AddChannel(ch *Channel) {
	if _, ok := n.chans[ch]; !ok {
		ch.nicks[n] = new(ChanPrivs)
		n.chans[ch] = ch.nicks[n]
	} else {
		logging.Warn("Nick.AddChannel(): trying to add already-present "+
			"channel %s to nick %s", ch.Name, n.Nick)
	}
}

// Returns true if the Nick is associated with the Channel.
func (n *Nick) IsOn(ch *Channel) bool {
	_, ok := n.chans[ch]
	return ok
}

// Returns true if the Nick is Me!
func (n *Nick) IsMe() bool {
	return n.me
}

// Disassociates a Channel from a Nick. Will call n.Delete() if the Nick is no
// longer on any channels we are tracking. Will also call ch.DelNick(n) to
// remove the association from the perspective of the Channel.
func (n *Nick) DelChannel(ch *Channel) {
	if _, ok := n.chans[ch]; ok {
		n.chans[ch] = nil, false
		ch.DelNick(n)
		if len(n.chans) == 0 {
			// nick is no longer in any channels we inhabit, stop tracking it
			n.Delete()
		}
	}
}

// Stops the Nick from being tracked by state tracking handlers. Also calls
// ch.DelNick(n) for all Nicks that are associated with the Channel.
func (n *Nick) Delete() {
	// we don't ever want to remove *our* nick from st.nicks...
	if !n.me {
		for ch, _ := range n.chans {
			ch.DelNick(n)
		}
		n.st.DelNick(n.Nick)
	}
}

// Parse mode strings for a Nick.
func (n *Nick) ParseModes(modes string) {
	var modeop bool // true => add mode, false => remove mode
	for i := 0; i < len(modes); i++ {
		switch m := modes[i]; m {
		case '+':
			modeop = true
		case '-':
			modeop = false
		case 'i':
			n.Modes.Invisible = modeop
		case 'o':
			n.Modes.Oper = modeop
		case 'w':
			n.Modes.WallOps = modeop
		case 'x':
			n.Modes.HiddenHost = modeop
		case 'z':
			n.Modes.SSL = modeop
		}
	}
}

// Returns a string representing the nick. Looks like:
//	Nick: <nick name> e.g. CowMaster
//	Hostmask: <ident@host> e.g. moo@cows.org
//	Real Name: <real name> e.g. Steve "CowMaster" Bush
//	Modes: <nick modes> e.g. +z
//	Channels:
//		<channel>: <privs> e.g. #moo: +o
//		...
func (n *Nick) String() string {
	str := "Nick: " + n.Nick + "\n\t"
	str += "Hostmask: " + n.Ident + "@" + n.Host + "\n\t"
	str += "Real Name: " + n.Name + "\n\t"
	str += "Modes: " + n.Modes.String() + "\n\t"
	if n.me {
		str += "I think this is ME!\n\t"
	}
	str += "Channels: \n"
	for ch, p := range n.chans {
		str += "\t\t" + ch.Name + ": " + p.String() + "\n"
	}
	return str
}

// Returns a string representing the nick modes. Looks like:
//	+iwx
func (nm *NickMode) String() string {
	str := "+"
	v := reflect.Indirect(reflect.ValueOf(nm))
	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		switch f := v.Field(i); f.Kind() {
		// only bools here at the mo!
		case reflect.Bool:
			if f.Bool() {
				str += NickModeToString[t.Field(i).Name]
			}
		}
	}
	if str == "+" {
		str = "No modes set"
	}
	return str
}
