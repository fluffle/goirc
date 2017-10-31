package state

import (
	"github.com/fluffle/goirc/logging"

	"sync"
)

// The state manager interface
type Tracker interface {
	// Nick methods
	NewNick(nick string) *Nick
	GetNick(nick string) *Nick
	ReNick(old, neu string) *Nick
	DelNick(nick string) *Nick
	NickInfo(nick, ident, host, name string) *Nick
	NickModes(nick, modestr string) *Nick
	// Channel methods
	NewChannel(channel string) *Channel
	GetChannel(channel string) *Channel
	DelChannel(channel string) *Channel
	Topic(channel, topic string) *Channel
	ChannelModes(channel, modestr string, modeargs ...string) *Channel
	// Information about ME!
	Me() *Nick
	// And the tracking operations
	IsOn(channel, nick string) (*ChanPrivs, bool)
	Associate(channel, nick string) *ChanPrivs
	Dissociate(channel, nick string)
	Wipe()
	// The state tracker can output a debugging string
	String() string
}

// ... and a struct to implement it ...
type stateTracker struct {
	// Map of channels we're on
	chans map[string]*channel
	// Map of nicks we know about
	nicks map[string]*nick

	// We need to keep state on who we are :-)
	me *nick

	// And we need to protect against data races *cough*.
	mu sync.Mutex
}

var _ Tracker = (*stateTracker)(nil)

// ... and a constructor to make it ...
func NewTracker(mynick string) *stateTracker {
	st := &stateTracker{
		chans: make(map[string]*channel),
		nicks: make(map[string]*nick),
	}
	st.me = newNick(mynick)
	st.nicks[mynick] = st.me
	return st
}

// ... and a method to wipe the state clean.
func (st *stateTracker) Wipe() {
	st.mu.Lock()
	defer st.mu.Unlock()
	// Deleting all the channels implicitly deletes every nick but me.
	for _, ch := range st.chans {
		st.delChannel(ch)
	}
}

/******************************************************************************\
 * tracker methods to create/look up nicks/channels
\******************************************************************************/

// Creates a new nick, initialises it, and stores it so it
// can be properly tracked for state management purposes.
func (st *stateTracker) NewNick(n string) *Nick {
	if n == "" {
		logging.Warn("Tracker.NewNick(): Not tracking empty nick.")
		return nil
	}
	st.mu.Lock()
	defer st.mu.Unlock()
	if _, ok := st.nicks[n]; ok {
		logging.Warn("Tracker.NewNick(): %s already tracked.", n)
		return nil
	}
	st.nicks[n] = newNick(n)
	return st.nicks[n].Nick()
}

// Returns a nick for the nick n, if we're tracking it.
func (st *stateTracker) GetNick(n string) *Nick {
	st.mu.Lock()
	defer st.mu.Unlock()
	if nk, ok := st.nicks[n]; ok {
		return nk.Nick()
	}
	return nil
}

// Signals to the tracker that a nick should be tracked
// under a "neu" nick rather than the old one.
func (st *stateTracker) ReNick(old, neu string) *Nick {
	st.mu.Lock()
	defer st.mu.Unlock()
	nk, ok := st.nicks[old]
	if !ok {
		logging.Warn("Tracker.ReNick(): %s not tracked.", old)
		return nil
	}
	if _, ok := st.nicks[neu]; ok {
		logging.Warn("Tracker.ReNick(): %s already exists.", neu)
		return nil
	}

	nk.nick = neu
	delete(st.nicks, old)
	st.nicks[neu] = nk
	for ch, _ := range nk.chans {
		// We also need to update the lookup maps of all the channels
		// the nick is on, to keep things in sync.
		delete(ch.lookup, old)
		ch.lookup[neu] = nk
	}
	return nk.Nick()
}

// Removes a nick from being tracked.
func (st *stateTracker) DelNick(n string) *Nick {
	st.mu.Lock()
	defer st.mu.Unlock()
	if nk, ok := st.nicks[n]; ok {
		if nk == st.me {
			logging.Warn("Tracker.DelNick(): won't delete myself.")
			return nil
		}
		st.delNick(nk)
		return nk.Nick()
	}
	logging.Warn("Tracker.DelNick(): %s not tracked.", n)
	return nil
}

func (st *stateTracker) delNick(nk *nick) {
	// st.mu lock held by DelNick, DelChannel or Wipe
	if nk == st.me {
		// Shouldn't get here => internal state tracking code is fubar.
		logging.Error("Tracker.DelNick(): TRYING TO DELETE ME :-(")
		return
	}
	delete(st.nicks, nk.nick)
	for ch, _ := range nk.chans {
		nk.delChannel(ch)
		ch.delNick(nk)
		if len(ch.nicks) == 0 {
			// Deleting a nick from tracking shouldn't empty any channels as
			// *we* should be on the channel with them to be tracking them.
			logging.Error("Tracker.delNick(): deleting nick %s emptied "+
				"channel %s, this shouldn't happen!", nk.nick, ch.name)
		}
	}
}

// Sets ident, host and "real" name for the nick.
func (st *stateTracker) NickInfo(n, ident, host, name string) *Nick {
	st.mu.Lock()
	defer st.mu.Unlock()
	nk, ok := st.nicks[n]
	if !ok {
		return nil
	}
	nk.ident = ident
	nk.host = host
	nk.name = name
	return nk.Nick()
}

// Sets user modes for the nick.
func (st *stateTracker) NickModes(n, modes string) *Nick {
	st.mu.Lock()
	defer st.mu.Unlock()
	nk, ok := st.nicks[n]
	if !ok {
		return nil
	}
	nk.parseModes(modes)
	return nk.Nick()
}

// Creates a new Channel, initialises it, and stores it so it
// can be properly tracked for state management purposes.
func (st *stateTracker) NewChannel(c string) *Channel {
	if c == "" {
		logging.Warn("Tracker.NewChannel(): Not tracking empty channel.")
		return nil
	}
	st.mu.Lock()
	defer st.mu.Unlock()
	if _, ok := st.chans[c]; ok {
		logging.Warn("Tracker.NewChannel(): %s already tracked.", c)
		return nil
	}
	st.chans[c] = newChannel(c)
	return st.chans[c].Channel()
}

// Returns a Channel for the channel c, if we're tracking it.
func (st *stateTracker) GetChannel(c string) *Channel {
	st.mu.Lock()
	defer st.mu.Unlock()
	if ch, ok := st.chans[c]; ok {
		return ch.Channel()
	}
	return nil
}

// Removes a Channel from being tracked.
func (st *stateTracker) DelChannel(c string) *Channel {
	st.mu.Lock()
	defer st.mu.Unlock()
	if ch, ok := st.chans[c]; ok {
		st.delChannel(ch)
		return ch.Channel()
	}
	logging.Warn("Tracker.DelChannel(): %s not tracked.", c)
	return nil
}

func (st *stateTracker) delChannel(ch *channel) {
	// st.mu lock held by DelChannel or Wipe
	delete(st.chans, ch.name)
	for nk, _ := range ch.nicks {
		ch.delNick(nk)
		nk.delChannel(ch)
		if len(nk.chans) == 0 && nk != st.me {
			// We're no longer in any channels with this nick.
			st.delNick(nk)
		}
	}
}

// Sets the topic of a channel.
func (st *stateTracker) Topic(c, topic string) *Channel {
	st.mu.Lock()
	defer st.mu.Unlock()
	ch, ok := st.chans[c]
	if !ok {
		return nil
	}
	ch.topic = topic
	return ch.Channel()
}

// Sets modes for a channel, including privileges like +o.
func (st *stateTracker) ChannelModes(c, modes string, args ...string) *Channel {
	st.mu.Lock()
	defer st.mu.Unlock()
	ch, ok := st.chans[c]
	if !ok {
		return nil
	}
	ch.parseModes(modes, args...)
	return ch.Channel()
}

// Returns the Nick the state tracker thinks is Me.
// NOTE: Nick() requires the mutex to be held.
func (st *stateTracker) Me() *Nick {
	st.mu.Lock()
	defer st.mu.Unlock()
	return st.me.Nick()
}

// Returns true if both the channel c and the nick n are tracked
// and the nick is associated with the channel.
func (st *stateTracker) IsOn(c, n string) (*ChanPrivs, bool) {
	st.mu.Lock()
	defer st.mu.Unlock()
	nk, nok := st.nicks[n]
	ch, cok := st.chans[c]
	if nok && cok {
		return nk.isOn(ch)
	}
	return nil, false
}

// Associates an already known nick with an already known channel.
func (st *stateTracker) Associate(c, n string) *ChanPrivs {
	st.mu.Lock()
	defer st.mu.Unlock()
	nk, nok := st.nicks[n]
	ch, cok := st.chans[c]

	if !cok {
		// As we can implicitly delete both nicks and channels from being
		// tracked by dissociating one from the other, we should verify that
		// we're not being passed an old Nick or Channel.
		logging.Error("Tracker.Associate(): channel %s not found in "+
			"internal state.", c)
		return nil
	} else if !nok {
		logging.Error("Tracker.Associate(): nick %s not found in "+
			"internal state.", n)
		return nil
	} else if _, ok := nk.isOn(ch); ok {
		logging.Warn("Tracker.Associate(): %s already on %s.",
			nk, ch)
		return nil
	}
	cp := new(ChanPrivs)
	ch.addNick(nk, cp)
	nk.addChannel(ch, cp)
	return cp.Copy()
}

// Dissociates an already known nick from an already known channel.
// Does some tidying up to stop tracking nicks we're no longer on
// any common channels with, and channels we're no longer on.
func (st *stateTracker) Dissociate(c, n string) {
	st.mu.Lock()
	defer st.mu.Unlock()
	nk, nok := st.nicks[n]
	ch, cok := st.chans[c]

	if !cok {
		// As we can implicitly delete both nicks and channels from being
		// tracked by dissociating one from the other, we should verify that
		// we're not being passed an old Nick or Channel.
		logging.Error("Tracker.Dissociate(): channel %s not found in "+
			"internal state.", c)
	} else if !nok {
		logging.Error("Tracker.Dissociate(): nick %s not found in "+
			"internal state.", n)
	} else if _, ok := nk.isOn(ch); !ok {
		logging.Warn("Tracker.Dissociate(): %s not on %s.",
			nk.nick, ch.name)
	} else if nk == st.me {
		// I'm leaving the channel for some reason, so it won't be tracked.
		st.delChannel(ch)
	} else {
		// Remove the nick from the channel and the channel from the nick.
		ch.delNick(nk)
		nk.delChannel(ch)
		if len(nk.chans) == 0 {
			// We're no longer in any channels with this nick.
			st.delNick(nk)
		}
	}
}

func (st *stateTracker) String() string {
	st.mu.Lock()
	defer st.mu.Unlock()
	str := "GoIRC Channels\n"
	str += "--------------\n\n"
	for _, ch := range st.chans {
		str += ch.String() + "\n"
	}
	str += "GoIRC NickNames\n"
	str += "---------------\n\n"
	for _, n := range st.nicks {
		if n != st.me {
			str += n.String() + "\n"
		}
	}
	return str
}
