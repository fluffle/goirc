package state

import (
	"fmt"
	"github.com/fluffle/goirc/logging"
	"reflect"
	"strconv"
)

// A struct representing an IRC channel
type Channel struct {
	Name, Topic string
	Modes       *ChanMode
	nicks       map[*Nick]*ChanPrivs
	st          StateTracker
}

// A struct representing the modes of an IRC Channel
// (the ones we care about, at least).
type ChanMode struct {
	// MODE +p, +s, +t, +n, +m
	Private, Secret, ProtectedTopic, NoExternalMsg, Moderated bool

	// MODE +i, +O, +z
	InviteOnly, OperOnly, SSLOnly bool

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

func (ch *Channel) initialise() {
	ch.Modes = new(ChanMode)
	ch.nicks = make(map[*Nick]*ChanPrivs)
}

// Associates a Nick with a Channel using a shared set of ChanPrivs
func (ch *Channel) AddNick(n *Nick) {
	if _, ok := ch.nicks[n]; !ok {
		ch.nicks[n] = new(ChanPrivs)
		n.chans[ch] = ch.nicks[n]
	} else {
		logging.Warn("Channel.AddNick(): trying to add already-present "+
			"nick %s to channel %s", n.Nick, ch.Name)
	}
}

// Returns true if the Nick is associated with the Channel
func (ch *Channel) IsOn(n *Nick) bool {
	_, ok := ch.nicks[n]
	return ok
}

// Disassociates a Nick from a Channel. Will call ch.Delete() if the Nick being
// removed is the connection's nick. Will also call n.DelChannel(ch) to remove
// the association from the perspective of the Nick.
func (ch *Channel) DelNick(n *Nick) {
	if _, ok := ch.nicks[n]; ok {
		if n.me {
			// we're leaving the channel, so remove all state we have about it
			ch.Delete()
		} else {
			ch.nicks[n] = nil, false
			n.DelChannel(ch)
		}
	}
	// we call Channel.DelNick() and Nick.DelChannel() from each other to ensure
	// consistency, and this would mean spewing an error message every delete
}

// Stops the Channel from being tracked by state tracking handlers. Also calls
// n.DelChannel(ch) for all nicks that are associated with the channel.
func (ch *Channel) Delete() {
	for n, _ := range ch.nicks {
		n.DelChannel(ch)
	}
	ch.st.DelChannel(ch.Name)
}

// Parses mode strings for a channel.
func (ch *Channel) ParseModes(modes string, modeargs []string) {
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
			ch.Modes.InviteOnly = modeop
		case 'm':
			ch.Modes.Moderated = modeop
		case 'n':
			ch.Modes.NoExternalMsg = modeop
		case 'p':
			ch.Modes.Private = modeop
		case 's':
			ch.Modes.Secret = modeop
		case 't':
			ch.Modes.ProtectedTopic = modeop
		case 'z':
			ch.Modes.SSLOnly = modeop
		case 'O':
			ch.Modes.OperOnly = modeop
		case 'k':
			if modeop && len(modeargs) != 0 {
				ch.Modes.Key, modeargs = modeargs[0], modeargs[1:]
			} else if !modeop {
				ch.Modes.Key = ""
			} else {
				logging.Warn("Channel.ParseModes(): not enough arguments to "+
					"process MODE %s %s%s", ch.Name, modestr, m)
			}
		case 'l':
			if modeop && len(modeargs) != 0 {
				ch.Modes.Limit, _ = strconv.Atoi(modeargs[0])
				modeargs = modeargs[1:]
			} else if !modeop {
				ch.Modes.Limit = 0
			} else {
				logging.Warn("Channel.ParseModes(): not enough arguments to "+
					"process MODE %s %s%s", ch.Name, modestr, m)
			}
		case 'q', 'a', 'o', 'h', 'v':
			if len(modeargs) != 0 {
				n := ch.st.GetNick(modeargs[0])
				if p, ok := ch.nicks[n]; ok {
					switch m {
					case 'q':
						p.Owner = modeop
					case 'a':
						p.Admin = modeop
					case 'o':
						p.Op = modeop
					case 'h':
						p.HalfOp = modeop
					case 'v':
						p.Voice = modeop
					}
					modeargs = modeargs[1:]
				} else {
					logging.Warn("Channel.ParseModes(): untracked nick %s "+
						"recieved MODE on channel %s", modeargs[0], ch.Name)
				}
			} else {
				logging.Warn("Channel.ParseModes(): not enough arguments to "+
					"process MODE %s %s%s", ch.Name, modestr, m)
			}
		}
	}
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
	for n, p := range ch.nicks {
		str += "\t\t" + n.Nick + ": " + p.String() + "\n"
	}
	return str
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
				a = append(a, fmt.Sprintf("%d", f.Int()))
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
func (p *ChanPrivs) String() string {
	str := "+"
	v := reflect.Indirect(reflect.ValueOf(p))
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
