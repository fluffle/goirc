package irc

// Here you'll find the Channel and Nick structs
// as well as the internal state maintenance code for the handlers

import (
	"fmt"
	"reflect"
)

// A struct representing an IRC channel
type Channel struct {
	Name, Topic string
	Modes       *ChanMode
	Nicks map[*Nick]*ChanPrivs
	conn  *Conn
}

// A struct representing an IRC nick
type Nick struct {
	Nick, Ident, Host, Name string
	Modes                   *NickMode
	Channels                map[*Channel]*ChanPrivs
	conn                    *Conn
}

// A struct representing the modes of an IRC Channel
// (the ones we care about, at least)
// see the MODE handler in setupEvents() for details
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
// This is only really useful for conn.Me, as we can't see other people's modes
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
 * Conn methods to create/look up nicks/channels
\******************************************************************************/
func (conn *Conn) NewNick(nick, ident, name, host string) *Nick {
	n := &Nick{Nick: nick, Ident: ident, Name: name, Host: host, conn: conn}
	n.initialise()
	conn.nicks[n.Nick] = n
	return n
}

func (conn *Conn) GetNick(n string) *Nick {
	if nick, ok := conn.nicks[n]; ok {
		return nick
	}
	return nil
}

func (conn *Conn) NewChannel(c string) *Channel {
	ch := &Channel{Name: c, conn: conn}
	ch.initialise()
	conn.chans[ch.Name] = ch
	return ch
}

func (conn *Conn) GetChannel(c string) *Channel {
	if ch, ok := conn.chans[c]; ok {
		return ch
	}
	return nil
}

/******************************************************************************\
 * Channel methods for state management
\******************************************************************************/
func (ch *Channel) initialise() {
	ch.Modes = new(ChanMode)
	ch.Nicks = make(map[*Nick]*ChanPrivs)
}

func (ch *Channel) AddNick(n *Nick) {
	if _, ok := ch.Nicks[n]; !ok {
		ch.Nicks[n] = new(ChanPrivs)
		n.Channels[ch] = ch.Nicks[n]
	} else {
		ch.conn.error("irc.Channel.AddNick() warning: trying to add already-present nick %s to channel %s", n.Nick, ch.Name)
	}
}

func (ch *Channel) DelNick(n *Nick) {
	if _, ok := ch.Nicks[n]; ok {
		fmt.Printf("irc.Channel.DelNick(): deleting %s from %s\n", n.Nick, ch.Name)
		if n == n.conn.Me {
			// we're leaving the channel, so remove all state we have about it
			ch.Delete()
		} else {
			ch.Nicks[n] = nil, false
			n.DelChannel(ch)
		}
	} // no else here ...
	// we call Channel.DelNick() and Nick.DelChan() from each other to ensure
	// consistency, and this would mean spewing an error message every delete
}

func (ch *Channel) Delete() {
	fmt.Printf("irc.Channel.Delete(): deleting %s\n", ch.Name)
	for n, _ := range ch.Nicks {
		n.DelChannel(ch)
	}
	ch.conn.chans[ch.Name] = nil, false
}

/******************************************************************************\
 * Nick methods for state management
\******************************************************************************/
func (n *Nick) initialise() {
	n.Modes = new(NickMode)
	n.Channels = make(map[*Channel]*ChanPrivs)
}

// very slightly different to Channel.AddNick() ...
func (n *Nick) AddChannel(ch *Channel) {
	if _, ok := n.Channels[ch]; !ok {
		ch.Nicks[n] = new(ChanPrivs)
		n.Channels[ch] = ch.Nicks[n]
	} else {
		n.conn.error("irc.Nick.AddChannel() warning: trying to add already-present channel %s to nick %s", ch.Name, n.Nick)
	}
}

func (n *Nick) DelChannel(ch *Channel) {
	if _, ok := n.Channels[ch]; ok {
		fmt.Printf("irc.Nick.DelChannel(): deleting %s from %s\n", n.Nick, ch.Name)
		n.Channels[ch] = nil, false
		ch.DelNick(n)
		if len(n.Channels) == 0 {
			// nick is no longer in any channels we inhabit, stop tracking it
			n.Delete()
		}
	}
}

func (n *Nick) ReNick(neu string) {
	n.conn.nicks[n.Nick] = nil, false
	n.Nick = neu
	n.conn.nicks[n.Nick] = n
}

func (n *Nick) Delete() {
	// we don't ever want to remove *our* nick from conn.nicks...
	if n != n.conn.Me {
		fmt.Printf("irc.Nick.Delete(): deleting %s\n", n.Nick)
		for ch, _ := range n.Channels {
			ch.DelNick(n)
		}
		n.conn.nicks[n.Nick] = nil, false
	}
}

/******************************************************************************\
 * String() methods for all structs in this file for ease of debugging.
\******************************************************************************/
var ChanModeToString = map[string]string{
	"Private": "p",
	"Secret": "s",
	"ProtectedTopic": "t",
	"NoExternalMsg": "n",
	"Moderated": "m",
	"InviteOnly": "i",
	"OperOnly": "O",
	"SSLOnly": "z",
	"Key": "k",
	"Limit": "l",
}
var NickModeToString = map[string]string{
	"Invisible": "i",
	"Oper": "o",
	"WallOps": "w",
	"HiddenHost": "x",
	"SSL": "z",
}
var ChanPrivToString = map[string]string{
	"Owner": "q",
	"Admin": "a",
	"Op": "o",
	"HalfOp": "h",
	"Voice": "v",
}
var ChanPrivToModeChar = map[string]byte{
	"Owner": '~',
	"Admin": '&',
	"Op": '@',
	"HalfOp": '%',
	"Voice": '+',
}
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

func (ch *Channel) String() string {
	str := "Channel: " + ch.Name + "\n\t"
	str += "Topic: " + ch.Topic + "\n\t"
	str += "Modes: " + ch.Modes.String() + "\n\t"
	str += "Nicks: \n"
	for n, p := range ch.Nicks {
		str += "\t\t" + n.Nick + ": " + p.String() + "\n"
	}
	return str
}

func (n *Nick) String() string {
	str := "Nick: " + n.Nick + "\n\t"
	str += "Hostmask: " + n.Ident + "@" + n.Host + "\n\t"
	str += "Real Name: " + n.Name + "\n\t"
	str += "Modes: " + n.Modes.String() + "\n\t"
	str += "Channels: \n"
	for ch, p := range n.Channels {
		str += "\t\t" + ch.Name + ": " + p.String() + "\n"
	}
	return str
}

func (cm *ChanMode) String() string {
	str := "+"
	a := make([]string, 2)
	v := reflect.Indirect(reflect.NewValue(cm)).(*reflect.StructValue)
	t := v.Type().(*reflect.StructType)
	for i := 0; i < v.NumField(); i++ {
		switch f := v.Field(i).(type) {
		case *reflect.BoolValue:
			if f.Get() {
				str += ChanModeToString[t.Field(i).Name]
			}
		case *reflect.StringValue:
			if f.Get() != "" {
				str += ChanModeToString[t.Field(i).Name]
				a[0] = f.Get()
			}
		case *reflect.IntValue:
			if f.Get() != 0 {
				str += ChanModeToString[t.Field(i).Name]
				a[1] = fmt.Sprintf("%d", cm.Limit)
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

func (nm *NickMode) String() string {
	str := "+"
	v := reflect.Indirect(reflect.NewValue(nm)).(*reflect.StructValue)
	t := v.Type().(*reflect.StructType)
	for i := 0; i < v.NumField(); i++ {
		switch f := v.Field(i).(type) {
		// only bools here at the mo!
		case *reflect.BoolValue:
			if f.Get() {
				str += NickModeToString[t.Field(i).Name]
			}
		}
	}
	if str == "+" {
		str = "No modes set"
	}
	return str
}

func (p *ChanPrivs) String() string {
	str := "+"
	v := reflect.Indirect(reflect.NewValue(p)).(*reflect.StructValue)
	t := v.Type().(*reflect.StructType)
	for i := 0; i < v.NumField(); i++ {
		switch f := v.Field(i).(type) {
		// only bools here at the mo too!
		case *reflect.BoolValue:
			if f.Get() {
				str += ChanPrivToString[t.Field(i).Name]
			}
		}
	}
	if str == "+" {
		str = "No modes set"
	}
	return str
}
