package state

import (
	"github.com/fluffle/goirc/logging"
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

/******************************************************************************\
 * tracker methods to create/look up nicks/channels
\******************************************************************************/

func NewTracker() *stateTracker {
	st := &stateTracker{}
	st.initialise()
	return st
}

func (st *stateTracker) initialise() {
	st.nicks = make(map[string]*Nick)
	st.chans = make(map[string]*Channel)
}

// Creates a new Nick, initialises it, and stores it so it
// can be properly tracked for state management purposes.
func (st *stateTracker) NewNick(nick string) *Nick {
	if _, ok := st.nicks[nick]; ok {
		logging.Warn("StateTracker.NewNick(): %s already tracked.", nick)
		return nil
	}
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
		if _, ok := st.nicks[neu]; !ok {
			st.nicks[old] = nil, false
			n.Nick = neu
			st.nicks[neu] = n
		} else {
			logging.Warn("StateTracker.ReNick(): %s already exists.", neu)
		}
	} else {
		logging.Warn("StateTracker.ReNick(): %s not tracked.", old)
	}
}

// Removes a Nick from being tracked.
func (st *stateTracker) DelNick(n string) {
	if _, ok := st.nicks[n]; ok {
		st.nicks[n] = nil, false
	} else {
		logging.Warn("StateTracker.DelNick(): %s not tracked.", n)
	}
}

// Creates a new Channel, initialises it, and stores it so it
// can be properly tracked for state management purposes.
func (st *stateTracker) NewChannel(c string) *Channel {
	if _, ok := st.chans[c]; ok {
		logging.Warn("StateTracker.NewChannel(): %s already tracked", c)
		return nil
	}
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
