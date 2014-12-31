package state

import (
	"github.com/fluffle/goirc/logging"

	"reflect"
	"strconv"
)

// A Channel is returned from the state tracker and contains
// a copy of the channel state at a particular time.
type Channel struct {
	Name, Topic string
	Modes       *ChanMode
	Nicks       map[string]*ChanPrivs
}

// Internal bookkeeping struct for channels.
type channel struct {
	name, topic string
	modes       *ChanMode
	lookup      map[string]*nick
	nicks       map[*nick]*ChanPrivs
}

// A struct representing the modes of an IRC Channel
// (the ones we care about, at least).
// http://www.unrealircd.com/files/docs/unreal32docs.html#userchannelmodes
type ChanMode struct {
	// MODE +p, +s, +t, +n, +m
	Private, Secret, ProtectedTopic, NoExternalMsg, Moderated bool

	// MODE +i, +O, +z
	InviteOnly, OperOnly, SSLOnly bool

	// MODE +r, +Z
	Registered, AllSSL bool

	// MODE +k
	Key string

	// MODE +l
	Limit int
}

// A struct representing the modes a Nick can have on a Channel
type ChanPrivs struct {
	// MODE +q, +a, +o, +h, +v
	Owner, Admin, Op, HalfOp, Voice bool
}

// Map ChanMode fields to IRC mode characters
var StringToChanMode = map[string]string{}
var ChanModeToString = map[string]string{
	"Private":        "p",
	"Secret":         "s",
	"ProtectedTopic": "t",
	"NoExternalMsg":  "n",
	"Moderated":      "m",
	"InviteOnly":     "i",
	"OperOnly":       "O",
	"SSLOnly":        "z",
	"Registered":     "r",
	"AllSSL":         "Z",
	"Key":            "k",
	"Limit":          "l",
}

// Map *irc.ChanPrivs fields to IRC mode characters
var StringToChanPriv = map[string]string{}
var ChanPrivToString = map[string]string{
	"Owner":  "q",
	"Admin":  "a",
	"Op":     "o",
	"HalfOp": "h",
	"Voice":  "v",
}

// Map *irc.ChanPrivs fields to the symbols used to represent these modes
// in NAMES and WHOIS responses
var ModeCharToChanPriv = map[byte]string{}
var ChanPrivToModeChar = map[string]byte{
	"Owner":  '~',
	"Admin":  '&',
	"Op":     '@',
	"HalfOp": '%',
	"Voice":  '+',
}

// Init function to fill in reverse mappings for *toString constants.
func init() {
	for k, v := range ChanModeToString {
		StringToChanMode[v] = k
	}
	for k, v := range ChanPrivToString {
		StringToChanPriv[v] = k
	}
	for k, v := range ChanPrivToModeChar {
		ModeCharToChanPriv[v] = k
	}
}

/******************************************************************************\
 * Channel methods for state management
\******************************************************************************/

func newChannel(name string) *channel {
	return &channel{
		name:   name,
		modes:  new(ChanMode),
		nicks:  make(map[*nick]*ChanPrivs),
		lookup: make(map[string]*nick),
	}
}

// Returns a copy of the internal tracker channel state at this time.
// Relies on tracker-level locking for concurrent access.
func (ch *channel) Channel() *Channel {
	c := &Channel{
		Name:  ch.name,
		Topic: ch.topic,
		Modes: ch.modes.Copy(),
		Nicks: make(map[string]*ChanPrivs),
	}
	for n, cp := range ch.nicks {
		c.Nicks[n.nick] = cp.Copy()
	}
	return c
}

func (ch *channel) isOn(nk *nick) (*ChanPrivs, bool) {
	cp, ok := ch.nicks[nk]
	return cp.Copy(), ok
}

// Associates a Nick with a Channel
func (ch *channel) addNick(nk *nick, cp *ChanPrivs) {
	if _, ok := ch.nicks[nk]; !ok {
		ch.nicks[nk] = cp
		ch.lookup[nk.nick] = nk
	} else {
		logging.Warn("Channel.addNick(): %s already on %s.", nk.nick, ch.name)
	}
}

// Disassociates a Nick from a Channel.
func (ch *channel) delNick(nk *nick) {
	if _, ok := ch.nicks[nk]; ok {
		delete(ch.nicks, nk)
		delete(ch.lookup, nk.nick)
	} else {
		logging.Warn("Channel.delNick(): %s not on %s.", nk.nick, ch.name)
	}
}

// Parses mode strings for a channel.
func (ch *channel) parseModes(modes string, modeargs ...string) {
	var modeop bool // true => add mode, false => remove mode
	var modestr string
	for i := 0; i < len(modes); i++ {
		switch m := modes[i]; m {
		case '+':
			modeop = true
			modestr = string(m)
		case '-':
			modeop = false
			modestr = string(m)
		case 'i':
			ch.modes.InviteOnly = modeop
		case 'm':
			ch.modes.Moderated = modeop
		case 'n':
			ch.modes.NoExternalMsg = modeop
		case 'p':
			ch.modes.Private = modeop
		case 'r':
			ch.modes.Registered = modeop
		case 's':
			ch.modes.Secret = modeop
		case 't':
			ch.modes.ProtectedTopic = modeop
		case 'z':
			ch.modes.SSLOnly = modeop
		case 'Z':
			ch.modes.AllSSL = modeop
		case 'O':
			ch.modes.OperOnly = modeop
		case 'k':
			if modeop && len(modeargs) != 0 {
				ch.modes.Key, modeargs = modeargs[0], modeargs[1:]
			} else if !modeop {
				ch.modes.Key = ""
			} else {
				logging.Warn("Channel.ParseModes(): not enough arguments to "+
					"process MODE %s %s%c", ch.name, modestr, m)
			}
		case 'l':
			if modeop && len(modeargs) != 0 {
				ch.modes.Limit, _ = strconv.Atoi(modeargs[0])
				modeargs = modeargs[1:]
			} else if !modeop {
				ch.modes.Limit = 0
			} else {
				logging.Warn("Channel.ParseModes(): not enough arguments to "+
					"process MODE %s %s%c", ch.name, modestr, m)
			}
		case 'q', 'a', 'o', 'h', 'v':
			if len(modeargs) != 0 {
				if nk, ok := ch.lookup[modeargs[0]]; ok {
					cp := ch.nicks[nk]
					switch m {
					case 'q':
						cp.Owner = modeop
					case 'a':
						cp.Admin = modeop
					case 'o':
						cp.Op = modeop
					case 'h':
						cp.HalfOp = modeop
					case 'v':
						cp.Voice = modeop
					}
					modeargs = modeargs[1:]
				} else {
					logging.Warn("Channel.ParseModes(): untracked nick %s "+
						"received MODE on channel %s", modeargs[0], ch.name)
				}
			} else {
				logging.Warn("Channel.ParseModes(): not enough arguments to "+
					"process MODE %s %s%c", ch.name, modestr, m)
			}
		default:
			logging.Info("Channel.ParseModes(): unknown mode char %c", m)
		}
	}
}

// Returns true if the Nick is associated with the Channel
func (ch *Channel) IsOn(nk string) (*ChanPrivs, bool) {
	cp, ok := ch.Nicks[nk]
	return cp, ok
}

// Test Channel equality.
func (ch *Channel) Equals(other *Channel) bool {
	return reflect.DeepEqual(ch, other)
}

// Duplicates a ChanMode struct.
func (cm *ChanMode) Copy() *ChanMode {
	if cm == nil { return nil }
	c := *cm
	return &c
}

// Test ChanMode equality.
func (cm *ChanMode) Equals(other *ChanMode) bool {
	return reflect.DeepEqual(cm, other)
}

// Duplicates a ChanPrivs struct.
func (cp *ChanPrivs) Copy() *ChanPrivs {
	if cp == nil { return nil }
	c := *cp
	return &c
}

// Test ChanPrivs equality.
func (cp *ChanPrivs) Equals(other *ChanPrivs) bool {
	return reflect.DeepEqual(cp, other)
}

// Returns a string representing the channel. Looks like:
//	Channel: <channel name> e.g. #moo
//	Topic: <channel topic> e.g. Discussing the merits of cows!
//	Mode: <channel modes> e.g. +nsti
//	Nicks:
//		<nick>: <privs> e.g. CowMaster: +o
//		...
func (ch *Channel) String() string {
	str := "Channel: " + ch.Name + "\n\t"
	str += "Topic: " + ch.Topic + "\n\t"
	str += "Modes: " + ch.Modes.String() + "\n\t"
	str += "Nicks: \n"
	for nk, cp := range ch.Nicks {
		str += "\t\t" + nk + ": " + cp.String() + "\n"
	}
	return str
}

func (ch *channel) String() string {
	return ch.Channel().String()
}

// Returns a string representing the channel modes. Looks like:
//	+npk key
func (cm *ChanMode) String() string {
	str := "+"
	a := make([]string, 0)
	v := reflect.Indirect(reflect.ValueOf(cm))
	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		switch f := v.Field(i); f.Kind() {
		case reflect.Bool:
			if f.Bool() {
				str += ChanModeToString[t.Field(i).Name]
			}
		case reflect.String:
			if f.String() != "" {
				str += ChanModeToString[t.Field(i).Name]
				a = append(a, f.String())
			}
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			if f.Int() != 0 {
				str += ChanModeToString[t.Field(i).Name]
				a = append(a, strconv.FormatInt(f.Int(), 10))
			}
		}
	}
	for _, s := range a {
		if s != "" {
			str += " " + s
		}
	}
	if str == "+" {
		str = "No modes set"
	}
	return str
}

// Returns a string representing the channel privileges. Looks like:
//	+o
func (cp *ChanPrivs) String() string {
	str := "+"
	v := reflect.Indirect(reflect.ValueOf(cp))
	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		switch f := v.Field(i); f.Kind() {
		// only bools here at the mo too!
		case reflect.Bool:
			if f.Bool() {
				str += ChanPrivToString[t.Field(i).Name]
			}
		}
	}
	if str == "+" {
		str = "No modes set"
	}
	return str
}
