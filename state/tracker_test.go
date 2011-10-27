package state

import (
	"github.com/fluffle/goirc/logging"
	"testing"
)

func TestSTNewTracker(t *testing.T) {
	l, m := logging.NewMock()
	l.SetLogLevel(logging.Debug)

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
	l.SetLogLevel(logging.Debug)
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
	l.SetLogLevel(logging.Debug)
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
	l.SetLogLevel(logging.Debug)
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
	l.SetLogLevel(logging.Debug)
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
	// Create a new channel for testing purposes
	chan1 := NewChannel("#test1", l)
	st.chans["#test1"] = chan1
	
	// Associate both "my" nick and test1 with the channel
	p := new(ChanPrivs)
	chan1.addNick(st.me, p)
	st.me.addChannel(chan1, p)
	chan1.addNick(nick1, p)
	nick1.addChannel(chan1, p)

	// Test we have the expected starting state (at least vaguely
	if len(chan1.nicks) != 2 || len(st.me.chans) != 1 || len(nick1.chans) != 1 {
		t.Errorf("Bad initial state for test DelNick() channel dissociation.")
	}

	st.DelNick("test1")

	// Actual deletion tested above...
	if len(chan1.nicks) != 1 || len(st.me.chans) != 1 || len(nick1.chans) != 0 {
		t.Errorf("Deleting nick didn't dissociate correctly from channels.")
	}

	if _, ok := chan1.nicks[nick1]; ok {
		t.Errorf("Nick not removed from channel's nick map.")
	}
	if _, ok := chan1.lookup["test1"]; ok {
		t.Errorf("Nick not removed from channels's lookup map.")
	}
}

func TestSTNewChannel(t *testing.T) {
	l, m := logging.NewMock()
	l.SetLogLevel(logging.Debug)
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
	l.SetLogLevel(logging.Debug)
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
	l.SetLogLevel(logging.Debug)
	st := NewTracker("mynick", l)

	test1 := NewChannel("#test1", l)
	st.chans["#test1"] = test1

	st.DelChannel("#test1")
	m.CheckNothingWritten(t)

	if _, ok := st.chans["#test1"]; ok {
		t.Errorf("Channel test1 still exists after DelChannel.")
	}
	if len(st.chans) != 0 {
		t.Errorf("Channel list still contains chans after DelChannel.")
	}

	st.chans["#test1"] = test1

	st.DelChannel("#test2")
	m.CheckWrittenAtLevel(t, logging.Warn,
		"StateTracker.DelChannel(): #test2 not tracked.")

	if len(st.chans) != 1 {
		t.Errorf("DelChannel had unexpected side-effects.")
	}
}

func TestSTIsOn(t *testing.T) {
	l, m := logging.NewMock()
	l.SetLogLevel(logging.Debug)
	st := NewTracker("mynick", l)

	nick1 := NewNick("test1", l)
	st.nicks["test1"] = nick1
	chan1 := NewChannel("#test1", l)
	st.chans["#test1"] = chan1

	if st.IsOn("#test1", "test1") {
		t.Errorf("test1 is not on #test1 (yet)")
	}
	cp := new(ChanPrivs)
	chan1.addNick(nick1, cp)
	nick1.addChannel(chan1, cp)
	if !st.IsOn("#test1", "test1") {
		t.Errorf("test1 is on #test1 (now)")
	}
	m.CheckNothingWritten(t)
}
