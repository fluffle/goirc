package state

import (
	"github.com/fluffle/goirc/logging"
	"reflect"
)

// A struct representing an IRC nick
type Nick struct {
	Nick, Ident, Host, Name string
	Modes                   *NickMode
	lookup					map[string]*Channel
	chans                   map[*Channel]*ChanPrivs
	l						logging.Logger
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

func NewNick(n string, l logging.Logger) *Nick {
	return &Nick{
		Nick: n,
		Modes: new(NickMode),
		chans: make(map[*Channel]*ChanPrivs),
		lookup: make(map[string]*Channel),
		l: l,
	}
}

// Returns true if the Nick is associated with the Channel.
func (nk *Nick) IsOn(ch *Channel) (*ChanPrivs, bool) {
	cp, ok := nk.chans[ch]
	return cp, ok
}

func (nk *Nick) IsOnStr(c string) (*Channel, bool) {
	ch, ok := nk.lookup[c]
	return ch, ok
}

// Associates a Channel with a Nick.
func (nk *Nick) addChannel(ch *Channel, cp *ChanPrivs) {
	if _, ok := nk.chans[ch]; !ok {
		nk.chans[ch] = cp
		nk.lookup[ch.Name] = ch
	} else {
		nk.l.Warn("Nick.addChannel(): %s already on %s.", nk.Nick, ch.Name)
	}
}

// Disassociates a Channel from a Nick.
func (nk *Nick) delChannel(ch *Channel) {
	if _, ok := nk.chans[ch]; ok {
		nk.chans[ch] = nil, false
		nk.lookup[ch.Name] = nil, false
	} else {
		nk.l.Warn("Nick.delChannel(): %s not on %s.", nk.Nick, ch.Name)
	}
}

// Parse mode strings for a Nick.
func (nk *Nick) ParseModes(modes string) {
	var modeop bool // true => add mode, false => remove mode
	for i := 0; i < len(modes); i++ {
		switch m := modes[i]; m {
		case '+':
			modeop = true
		case '-':
			modeop = false
		case 'i':
			nk.Modes.Invisible = modeop
		case 'o':
			nk.Modes.Oper = modeop
		case 'w':
			nk.Modes.WallOps = modeop
		case 'x':
			nk.Modes.HiddenHost = modeop
		case 'z':
			nk.Modes.SSL = modeop
		default:
			nk.l.Info("Nick.ParseModes(): unknown mode char %c", m)
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
func (nk *Nick) String() string {
	str := "Nick: " + nk.Nick + "\n\t"
	str += "Hostmask: " + nk.Ident + "@" + nk.Host + "\n\t"
	str += "Real Name: " + nk.Name + "\n\t"
	str += "Modes: " + nk.Modes.String() + "\n\t"
	str += "Channels: \n"
	for ch, cp := range nk.chans {
		str += "\t\t" + ch.Name + ": " + cp.String() + "\n"
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
