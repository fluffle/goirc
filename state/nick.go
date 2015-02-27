package state

import (
	"github.com/fluffle/goirc/logging"

	"reflect"
)

// A Nick is returned from the state tracker and contains
// a copy of the nick state at a particular time.
type Nick struct {
	Nick, Ident, Host, Name string
	Modes                   *NickMode
	Channels                map[string]*ChanPrivs
}

// Internal bookkeeping struct for nicks.
type nick struct {
	nick, ident, host, name string
	modes                   *NickMode
	lookup                  map[string]*channel
	chans                   map[*channel]*ChanPrivs
}

// A struct representing the modes of an IRC Nick (User Modes)
// (again, only the ones we care about)
//
// This is only really useful for me, as we can't see other people's modes
// without IRC operator privileges (and even then only on some IRCd's).
type NickMode struct {
	// MODE +B, +i, +o, +w, +x, +z
	Bot, Invisible, Oper, WallOps, HiddenHost, SSL bool
}

// Map *irc.NickMode fields to IRC mode characters and vice versa
var StringToNickMode = map[string]string{}
var NickModeToString = map[string]string{
	"Bot":        "B",
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
 * nick methods for state management
\******************************************************************************/

func newNick(n string) *nick {
	return &nick{
		nick:   n,
		modes:  new(NickMode),
		chans:  make(map[*channel]*ChanPrivs),
		lookup: make(map[string]*channel),
	}
}

// Returns a copy of the internal tracker nick state at this time.
// Relies on tracker-level locking for concurrent access.
func (nk *nick) Nick() *Nick {
	n := &Nick{
		Nick:     nk.nick,
		Ident:    nk.ident,
		Host:     nk.host,
		Name:     nk.name,
		Modes:    nk.modes.Copy(),
		Channels: make(map[string]*ChanPrivs),
	}
	for c, cp := range nk.chans {
		n.Channels[c.name] = cp.Copy()
	}
	return n
}

func (nk *nick) isOn(ch *channel) (*ChanPrivs, bool) {
	cp, ok := nk.chans[ch]
	return cp.Copy(), ok
}

// Associates a Channel with a Nick.
func (nk *nick) addChannel(ch *channel, cp *ChanPrivs) {
	if _, ok := nk.chans[ch]; !ok {
		nk.chans[ch] = cp
		nk.lookup[ch.name] = ch
	} else {
		logging.Warn("Nick.addChannel(): %s already on %s.", nk.nick, ch.name)
	}
}

// Disassociates a Channel from a Nick.
func (nk *nick) delChannel(ch *channel) {
	if _, ok := nk.chans[ch]; ok {
		delete(nk.chans, ch)
		delete(nk.lookup, ch.name)
	} else {
		logging.Warn("Nick.delChannel(): %s not on %s.", nk.nick, ch.name)
	}
}

// Parse mode strings for a Nick.
func (nk *nick) parseModes(modes string) {
	var modeop bool // true => add mode, false => remove mode
	for i := 0; i < len(modes); i++ {
		switch m := modes[i]; m {
		case '+':
			modeop = true
		case '-':
			modeop = false
		case 'B':
			nk.modes.Bot = modeop
		case 'i':
			nk.modes.Invisible = modeop
		case 'o':
			nk.modes.Oper = modeop
		case 'w':
			nk.modes.WallOps = modeop
		case 'x':
			nk.modes.HiddenHost = modeop
		case 'z':
			nk.modes.SSL = modeop
		default:
			logging.Info("Nick.ParseModes(): unknown mode char %c", m)
		}
	}
}

// Returns true if the Nick is associated with the Channel.
func (nk *Nick) IsOn(ch string) (*ChanPrivs, bool) {
	cp, ok := nk.Channels[ch]
	return cp, ok
}

// Tests Nick equality.
func (nk *Nick) Equals(other *Nick) bool {
	return reflect.DeepEqual(nk, other)
}

// Duplicates a NickMode struct.
func (nm *NickMode) Copy() *NickMode {
	if nm == nil { return nil }
	n := *nm
	return &n
}

// Tests NickMode equality.
func (nm *NickMode) Equals(other *NickMode) bool {
	return reflect.DeepEqual(nm, other)
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
	for ch, cp := range nk.Channels {
		str += "\t\t" + ch + ": " + cp.String() + "\n"
	}
	return str
}

func (nk *nick) String() string {
	return nk.Nick().String()
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
