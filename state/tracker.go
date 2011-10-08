package state

// Here you'll find the Channel and Nick structs
// as well as the internal state maintenance code for the handlers

import (
	"fmt"
	"github.com/fluffle/goirc/logging"
	"reflect"
	"strconv"
)

// The state manager interface
type StateTracker interface {
	NewNick(nick string) *Nick
	GetNick(nick string) *Nick
	ReNick(old, neu string)
	DelNick(nick string)
	NewChannel(channel string) *Channel
	GetChannel(channel string) *Channel
	DelChannel(channel string)
	IsOn(channel, nick string) bool
}

// ... and a struct to implement it
type stateTracker struct {
	// Map of channels we're on
	chans map[string]*Channel
	// Map of nicks we know about
	nicks map[string]*Nick
}

// A struct representing an IRC channel
type Channel struct {
	Name, Topic string
	Modes       *ChanMode
	nicks       map[*Nick]*ChanPrivs
	st          StateTracker
}

// A struct representing an IRC nick
type Nick struct {
	Nick, Ident, Host, Name string
	Modes                   *NickMode
	chans                   map[*Channel]*ChanPrivs
	me                      bool
	st                      StateTracker
}

// A struct representing the modes of an IRC Channel
// (the ones we care about, at least).
//
// See the MODE handler in setupEvents() for details of how this is maintained.
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

// A struct representing the modes of an IRC Nick (User Modes)
// (again, only the ones we care about)
//
// This is only really useful for me, as we can't see other people's modes
// without IRC operator privileges (and even then only on some IRCd's).
type NickMode struct {
	// MODE +i, +o, +w, +x, +z
	Invisible, Oper, WallOps, HiddenHost, SSL bool
}

// A struct representing the modes a Nick can have on a Channel
type ChanPrivs struct {
	// MODE +q, +a, +o, +h, +v
	Owner, Admin, Op, HalfOp, Voice bool
}

/******************************************************************************\
 * tracker methods to create/look up nicks/channels
\******************************************************************************/

func NewTracker() StateTracker {
	st := &stateTracker{}
	st.initialise()
}

func (st *stateTracker) initialise() {
	st.nicks = make(map[string]*Nick)
	st.chans = make(map[string]*Channel)
}

// Creates a new Nick, initialises it, and stores it so it
// can be properly tracked for state management purposes.
func (st *stateTracker) NewNick(nick string) *Nick {
	n := &Nick{Nick: nick, st: st}
	n.initialise()
	st.nicks[nick] = n
	return n
}

// Returns a Nick for the nick n, if we're tracking it.
func (st *stateTracker) GetNick(n string) *Nick {
	if nick, ok := st.nicks[n]; ok {
		return nick
	}
	return nil
}

// Signals to the tracker that a Nick should be tracked
// under a "neu" nick rather than the old one.
func (st *stateTracker) ReNick(old, neu string) {
	if n, ok := st.nicks[old]; ok {
		st.nicks[old] = nil, false
		n.Nick = neu
		st.nicks[neu] = n
	}
}

// Removes a Nick from being tracked.
func (st *stateTracker) DelNick(n string) {
	if _, ok := st.nicks[n]; ok {
		st.nicks[n] = nil, false
	}
}

// Creates a new Channel, initialises it, and stores it so it
// can be properly tracked for state management purposes.
func (st *stateTracker) NewChannel(c string) *Channel {
	ch := &Channel{Name: c, st: st}
	ch.initialise()
	st.chans[c] = ch
	return ch
}

// Returns a Channel for the channel c, if we're tracking it.
func (st *stateTracker) GetChannel(c string) *Channel {
	if ch, ok := st.chans[c]; ok {
		return ch
	}
	return nil
}

// Removes a Channel from being tracked.
func (st *stateTracker) DelChannel(c string) {
	if _, ok := st.chans[c]; ok {
		st.chans[c] = nil, false
	}
}

// Returns true if both the channel c and the nick n are tracked
// and the nick is associated with the channel.
func (st *stateTracker) IsOn(c, n string) bool {
	nk := st.GetNick(n)
	ch := st.GetChannel(c)
	if nk != nil && ch != nil {
		return nk.IsOn(ch)
	}
	return false
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
	return nil
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

/******************************************************************************\
 * String() methods for all structs in this file for ease of debugging.
\******************************************************************************/

// Map ChanMode fields to IRC mode characters
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

// Map *irc.NickMode fields to IRC mode characters
var NickModeToString = map[string]string{
	"Invisible":  "i",
	"Oper":       "o",
	"WallOps":    "w",
	"HiddenHost": "x",
	"SSL":        "z",
}

// Map *irc.ChanPrivs fields to IRC mode characters
var ChanPrivToString = map[string]string{
	"Owner":  "q",
	"Admin":  "a",
	"Op":     "o",
	"HalfOp": "h",
	"Voice":  "v",
}

// Map *irc.ChanPrivs fields to the symbols used to represent these modes
// in NAMES and WHOIS responses
var ChanPrivToModeChar = map[string]byte{
	"Owner":  '~',
	"Admin":  '&',
	"Op":     '@',
	"HalfOp": '%',
	"Voice":  '+',
}

// Reverse mappings of the above datastructures
var StringToChanMode, StringToNickMode, StringToChanPriv map[string]string
var ModeCharToChanPriv map[byte]string

// Init function to fill in reverse mappings for *toString constants etc.
func init() {
	StringToChanMode = make(map[string]string)
	for k, v := range ChanModeToString {
		StringToChanMode[v] = k
	}
	StringToNickMode = make(map[string]string)
	for k, v := range NickModeToString {
		StringToNickMode[v] = k
	}
	StringToChanPriv = make(map[string]string)
	for k, v := range ChanPrivToString {
		StringToChanPriv[v] = k
	}
	ModeCharToChanPriv = make(map[byte]string)
	for k, v := range ChanPrivToModeChar {
		ModeCharToChanPriv[v] = k
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
