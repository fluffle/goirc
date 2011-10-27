package state

import (
	"github.com/fluffle/goirc/logging"
	"testing"
)

func TestSTNewTracker(t *testing.T) {
	l, m := logging.NewMock()

	st := NewTracker("mynick", l)
	m.CheckNothingWritten(t)

	if st.l != l {
		t.Errorf("State tracker's logger not set correctly.")
	}
	if len(st.nicks) != 1 {
		t.Errorf("Nick list of new tracker is not 1 (me!).")
	}
	if len(st.chans) != 0 {
		t.Errorf("Channel list of new tracker is not empty.")
	}
	if nk, ok := st.nicks["mynick"]; !ok || nk.Nick != "mynick" || nk != st.me {
		t.Errorf("My nick not stored correctly in tracker.")
	}
}

func TestSTNewNick(t *testing.T) {
	l, m := logging.NewMock()
	st := NewTracker("mynick", l)

	test1 := st.NewNick("test1")
	m.CheckNothingWritten(t)

	if test1 == nil || test1.Nick != "test1" || test1.l != l {
		t.Errorf("Nick object created incorrectly by NewNick.")
	}
	if n, ok := st.nicks["test1"]; !ok || n != test1 || len(st.nicks) != 2 {
		t.Errorf("Nick object stored incorrectly by NewNick.")
	}

	if fail := st.NewNick("test1"); fail != nil {
		t.Errorf("Creating duplicate nick did not produce nil return.")
	}
	m.CheckWrittenAtLevel(t, logging.Warn,
		"StateTracker.NewNick(): test1 already tracked.")
}

func TestSTGetNick(t *testing.T) {
	l, _ := logging.NewMock()
	st := NewTracker("mynick", l)

	test1 := NewNick("test1", l)
	st.nicks["test1"] = test1

	if n := st.GetNick("test1"); n != test1 {
		t.Errorf("Incorrect nick returned by GetNick.")
	}
	if n := st.GetNick("test2"); n != nil {
		t.Errorf("Nick unexpectedly returned by GetNick.")
	}
	if len(st.nicks) != 2 {
		t.Errorf("Nick list changed size during GetNick.")
	}
}

func TestSTReNick(t *testing.T) {
	l, m := logging.NewMock()
	st := NewTracker("mynick", l)

	test1 := NewNick("test1", l)
	st.nicks["test1"] = test1

	// This channel is here to ensure that its lookup map gets updated
	cp := new(ChanPrivs)
	chan1 := NewChannel("#chan1", l)
	test1.addChannel(chan1, cp)
	chan1.addNick(test1, cp)

	st.ReNick("test1", "test2")
	m.CheckNothingWritten(t)

	if _, ok := st.nicks["test1"]; ok {
		t.Errorf("Nick test1 still exists after ReNick.")
	}
	if n, ok := st.nicks["test2"]; !ok || n != test1 {
		t.Errorf("Nick test2 doesn't exist after ReNick.")
	}
	if _, ok := chan1.lookup["test1"]; ok {
		t.Errorf("Channel #chan1 still knows about test1 after ReNick.")
	}
	if n, ok := chan1.lookup["test2"]; !ok || n != test1 {
		t.Errorf("Channel #chan1 doesn't know about test2 after ReNick.")
	}
	if test1.Nick != "test2" {
		t.Errorf("Nick test1 not changed correctly.")
	}
	if len(st.nicks) != 2 {
		t.Errorf("Nick list changed size during ReNick.")
	}

	test2 := NewNick("test1", l)
	st.nicks["test1"] = test2

	st.ReNick("test1", "test2")
	m.CheckWrittenAtLevel(t, logging.Warn,
		"StateTracker.ReNick(): test2 already exists.")

	if n, ok := st.nicks["test2"]; !ok || n != test1 {
		t.Errorf("Nick test2 overwritten/deleted by ReNick.")
	}
	if n, ok := st.nicks["test1"]; !ok || n != test2 {
		t.Errorf("Nick test1 overwritten/deleted by ReNick.")
	}
	if len(st.nicks) != 3 {
		t.Errorf("Nick list changed size during ReNick.")
	}

	st.ReNick("test3", "test2")
	m.CheckWrittenAtLevel(t, logging.Warn,
		"StateTracker.ReNick(): test3 not tracked.")
}

func TestSTDelNick(t *testing.T) {
	l, m := logging.NewMock()
	st := NewTracker("mynick", l)

	nick1 := NewNick("test1", l)
	st.nicks["test1"] = nick1

	st.DelNick("test1")
	m.CheckNothingWritten(t)

	if _, ok := st.nicks["test1"]; ok {
		t.Errorf("Nick test1 still exists after DelNick.")
	}
	if len(st.nicks) != 1 {
		t.Errorf("Nick list still contains nicks after DelNick.")
	}

	st.nicks["test1"] = nick1

	st.DelNick("test2")
	m.CheckWrittenAtLevel(t, logging.Warn,
		"StateTracker.DelNick(): test2 not tracked.")

	if len(st.nicks) != 2 {
		t.Errorf("Deleting unknown nick had unexpected side-effects.")
	}

	// Deleting my nick shouldn't work
	st.DelNick("mynick")
	m.CheckWrittenAtLevel(t, logging.Warn,
		"StateTracker.DelNick(): won't delete myself.")

	if len(st.nicks) != 2 {
		t.Errorf("Deleting myself had unexpected side-effects.")
	}

	// Test that deletion correctly dissociates nick from channels.
	// NOTE: the two error states in delNick (as opposed to DelNick)
	// are not tested for here, as they will only arise from programming
	// errors in other methods. Using m.CheckNothingWritten() while
	// testing these methods will catch those errors should they occur.

	// Create a new channel for testing purposes
	chan1 := NewChannel("#test1", l)
	st.chans["#test1"] = chan1

	// Associate both "my" nick and test1 with the channel
	p := new(ChanPrivs)
	chan1.addNick(st.me, p)
	st.me.addChannel(chan1, p)
	chan1.addNick(nick1, p)
	nick1.addChannel(chan1, p)

	// Test we have the expected starting state (at least vaguely)
	if len(chan1.nicks) != 2 || len(st.nicks) != 2 ||
		len(st.me.chans) != 1 || len(nick1.chans) != 1 || len(st.chans) != 1 {
		t.Errorf("Bad initial state for test DelNick() channel dissociation.")
	}

	st.DelNick("test1")
	m.CheckNothingWritten(t)

	// Actual deletion tested above...
	if len(chan1.nicks) != 1 || len(st.chans) != 1 ||
		len(st.me.chans) != 1 || len(nick1.chans) != 0 || len(st.chans) != 1 {
		t.Errorf("Deleting nick didn't dissociate correctly from channels.")
	}

	if _, ok := chan1.nicks[nick1]; ok {
		t.Errorf("Nick not removed from channel's nick map.")
	}
	if _, ok := chan1.lookup["test1"]; ok {
		t.Errorf("Nick not removed from channel's lookup map.")
	}
}

func TestSTNewChannel(t *testing.T) {
	l, m := logging.NewMock()
	st := NewTracker("mynick", l)

	if len(st.chans) != 0 {
		t.Errorf("Channel list of new tracker is non-zero length.")
	}

	test1 := st.NewChannel("#test1")
	m.CheckNothingWritten(t)

	if test1 == nil || test1.Name != "#test1" || test1.l != l {
		t.Errorf("Channel object created incorrectly by NewChannel.")
	}
	if c, ok := st.chans["#test1"]; !ok || c != test1 || len(st.chans) != 1 {
		t.Errorf("Channel object stored incorrectly by NewChannel.")
	}

	if fail := st.NewChannel("#test1"); fail != nil {
		t.Errorf("Creating duplicate chan did not produce nil return.")
	}
	m.CheckWrittenAtLevel(t, logging.Warn,
		"StateTracker.NewChannel(): #test1 already tracked.")
}

func TestSTGetChannel(t *testing.T) {
	l, _ := logging.NewMock()
	st := NewTracker("mynick", l)

	test1 := NewChannel("#test1", l)
	st.chans["#test1"] = test1

	if c := st.GetChannel("#test1"); c != test1 {
		t.Errorf("Incorrect Channel returned by GetChannel.")
	}
	if c := st.GetChannel("#test2"); c != nil {
		t.Errorf("Channel unexpectedly returned by GetChannel.")
	}
	if len(st.chans) != 1 {
		t.Errorf("Channel list changed size during GetChannel.")
	}
}

func TestSTDelChannel(t *testing.T) {
	l, m := logging.NewMock()
	st := NewTracker("mynick", l)

	chan1 := NewChannel("#test1", l)
	st.chans["#test1"] = chan1

	st.DelChannel("#test1")
	m.CheckNothingWritten(t)

	if _, ok := st.chans["#test1"]; ok {
		t.Errorf("Channel test1 still exists after DelChannel.")
	}
	if len(st.chans) != 0 {
		t.Errorf("Channel list still contains chans after DelChannel.")
	}

	st.chans["#test1"] = chan1

	st.DelChannel("#test2")
	m.CheckWrittenAtLevel(t, logging.Warn,
		"StateTracker.DelChannel(): #test2 not tracked.")

	if len(st.chans) != 1 {
		t.Errorf("DelChannel had unexpected side-effects.")
	}

	// Test that deletion correctly dissociates channel from tracked nicks.

	// In order to test this thoroughly we need two channels (so that delNick()
	// is not called internally in delChannel() when len(nick1.chans) == 0.
	chan2 := NewChannel("#test2", l)
	st.chans["#test2"] = chan2
	nick1 := NewNick("test1", l)
	st.nicks["test1"] = nick1

	// Associate both "my" nick and test1 with the channels
	p := new(ChanPrivs)
	chan1.addNick(st.me, p)
	st.me.addChannel(chan1, p)
	chan1.addNick(nick1, p)
	nick1.addChannel(chan1, p)
	chan2.addNick(st.me, p)
	st.me.addChannel(chan2, p)
	chan2.addNick(nick1, p)
	nick1.addChannel(chan2, p)

	// Test we have the expected starting state (at least vaguely)
	if len(chan1.nicks) != 2 || len(chan2.nicks) != 2 || len(st.nicks) != 2 ||
		len(st.me.chans) != 2 || len(nick1.chans) != 2 || len(st.chans) != 2 {
		t.Errorf("Bad initial state for test DelChannel() nick dissociation.")
	}

	st.DelChannel("#test1")
	m.CheckNothingWritten(t)

	// Test intermediate state. We're still on #test2 with test1, so test1
	// shouldn't be deleted from state tracking itself just yet.
	if len(chan1.nicks) != 0 || len(chan2.nicks) != 2 || len(st.nicks) != 2 ||
		len(st.me.chans) != 1 || len(nick1.chans) != 1 || len(st.chans) != 1 {
		t.Errorf("Deleting channel didn't dissociate correctly from nicks.")
	}

	if _, ok := nick1.chans[chan1]; ok {
		t.Errorf("Channel not removed from nick's chans map.")
	}
	if _, ok := nick1.lookup["#test1"]; ok {
		t.Errorf("Channel not removed from nick's lookup map.")
	}

	st.DelChannel("#test2")
	m.CheckNothingWritten(t)

	// Test final state. Deleting #test2 means that we're no longer on any
	// common channels with test1, and thus it should be removed from tracking.
	if len(chan1.nicks) != 0 || len(chan2.nicks) != 0 || len(st.nicks) != 1 ||
		len(st.me.chans) != 0 || len(nick1.chans) != 0 || len(st.chans) != 0 {
		t.Errorf("Deleting last channel didn't dissociate correctly from nicks.")
	}

	if _, ok := st.nicks["test1"]; ok {
		t.Errorf("Nick not deleted correctly when on no channels.")
	}
	if _, ok := st.nicks["mynick"]; !ok {
		t.Errorf("My nick deleted incorrectly when on no channels.")
	}
}

func TestSTIsOn(t *testing.T) {
	l, m := logging.NewMock()
	st := NewTracker("mynick", l)

	nick1 := NewNick("test1", l)
	st.nicks["test1"] = nick1
	chan1 := NewChannel("#test1", l)
	st.chans["#test1"] = chan1

	if priv, ok := st.IsOn("#test1", "test1"); ok || priv != nil {
		t.Errorf("test1 is not on #test1 (yet)")
	}
	cp := new(ChanPrivs)
	chan1.addNick(nick1, cp)
	nick1.addChannel(chan1, cp)
	if priv, ok := st.IsOn("#test1", "test1"); !ok || priv != cp {
		t.Errorf("test1 is on #test1 (now)")
	}
	m.CheckNothingWritten(t)
}

func TestSTAssociate(t *testing.T) {
	l, m := logging.NewMock()
	st := NewTracker("mynick", l)

	nick1 := st.NewNick("test1")
	chan1 := st.NewChannel("#test1")

	cp := st.Associate(chan1, nick1)
	m.CheckNothingWritten(t)
	if priv, ok := st.IsOn("#test1", "test1"); !ok || cp != priv {
		t.Errorf("test1 was not associated with #test1.")
	}

	// Test error cases
	st.Associate(nil, nick1)
	m.CheckWrittenAtLevel(t, logging.Error,
		"StateTracker.Associate(): passed nil values :-(")

	st.Associate(chan1, nil)
	m.CheckWrittenAtLevel(t, logging.Error,
		"StateTracker.Associate(): passed nil values :-(")

	st.Associate(chan1, nick1)
	m.CheckWrittenAtLevel(t, logging.Warn,
		"StateTracker.Associate(): test1 already on #test1.")

	nick2 := NewNick("test2", l)
	st.Associate(chan1, nick2)
	m.CheckWrittenAtLevel(t, logging.Error,
		"StateTracker.Associate(): nick test2 not found in (or differs from) "+
		"internal state.")

	chan2 := NewChannel("#test2", l)
	st.Associate(chan2, nick1)
	m.CheckWrittenAtLevel(t, logging.Error,
		"StateTracker.Associate(): channel #test2 not found in (or differs "+
		"from) internal state.")
}

func TestSTDissociate(t *testing.T) {
	l, m := logging.NewMock()
	st := NewTracker("mynick", l)

	nick1 := st.NewNick("test1")
	chan1 := st.NewChannel("#test1")
	chan2 := st.NewChannel("#test2")

	// Associate both "my" nick and test1 with the channels
	st.Associate(chan1, st.me)
	st.Associate(chan1, nick1)
	st.Associate(chan2, st.me)
	st.Associate(chan2, nick1)

	// Check the initial state looks mostly like we expect it to.
	if len(chan1.nicks) != 2 || len(chan2.nicks) != 2 || len(st.nicks) != 2 ||
		len(st.me.chans) != 2 || len(nick1.chans) != 2 || len(st.chans) != 2 {
		t.Errorf("Initial state for dissociation tests looks odd.")
	}

	// First, test the case of me leaving #test2
	st.Dissociate(chan2, st.me)
	m.CheckNothingWritten(t)

	// This should have resulted in the complete deletion of the channel.
	if len(chan1.nicks) != 2 || len(chan2.nicks) != 0 || len(st.nicks) != 2 ||
		len(st.me.chans) != 1 || len(nick1.chans) != 1 || len(st.chans) != 1 {
		t.Errorf("Dissociating myself from channel didn't delete it correctly.")
	}

	// Reassociating myself and test1 to #test2 shouldn't cause any errors.
	chan2 = st.NewChannel("#test2")
	st.Associate(chan2, st.me)
	st.Associate(chan2, nick1)
	m.CheckNothingWritten(t)

	// Check state once moar.
	if len(chan1.nicks) != 2 || len(chan2.nicks) != 2 || len(st.nicks) != 2 ||
		len(st.me.chans) != 2 || len(nick1.chans) != 2 || len(st.chans) != 2 {
		t.Errorf("Reassociating to channel has produced unexpected state.")
	}

	// Now, lets dissociate test1 from #test1 then #test2.
	// This first one should only result in a change in associations.
	st.Dissociate(chan1, nick1)
	m.CheckNothingWritten(t)

	if len(chan1.nicks) != 1 || len(chan2.nicks) != 2 || len(st.nicks) != 2 ||
		len(st.me.chans) != 2 || len(nick1.chans) != 1 || len(st.chans) != 2 {
		t.Errorf("Dissociating a nick from one channel went wrong.")
	}

	// This second one should also delete test1
	// as it's no longer on any common channels with us
	st.Dissociate(chan2, nick1)
	m.CheckNothingWritten(t)

	if len(chan1.nicks) != 1 || len(chan2.nicks) != 1 || len(st.nicks) != 1 ||
		len(st.me.chans) != 2 || len(nick1.chans) != 0 || len(st.chans) != 2 {
		t.Errorf("Dissociating a nick from it's last channel went wrong.")
	}

	// Check error cases
	// test1 was deleted above, so "re-track" it for this test.
	nick1 = st.NewNick("test1")
	st.Dissociate(chan1, nick1)
	m.CheckWrittenAtLevel(t, logging.Warn,
		"StateTracker.Dissociate(): test1 not on #test1.")

	st.Dissociate(chan1, nil)
	m.CheckWrittenAtLevel(t, logging.Error,
		"StateTracker.Dissociate(): passed nil values :-(")

	st.Dissociate(nil, nick1)
	m.CheckWrittenAtLevel(t, logging.Error,
		"StateTracker.Dissociate(): passed nil values :-(")

	nick3 := NewNick("test3", l)
	st.Dissociate(chan1, nick3)
	m.CheckWrittenAtLevel(t, logging.Error,
		"StateTracker.Dissociate(): nick test3 not found in (or differs from) "+
		"internal state.")

	chan3 := NewChannel("#test3", l)
	st.Dissociate(chan3, nick1)
	m.CheckWrittenAtLevel(t, logging.Error,
		"StateTracker.Dissociate(): channel #test3 not found in (or differs "+
		"from) internal state.")
}
