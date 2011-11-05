package state

import (
	"testing"
)

func TestNewChannel(t *testing.T) {
	_, s := setUp(t)
	defer s.tearDown()

	ch := NewChannel("#test1", s.log)

	if ch.Name != "#test1" || ch.l != s.log {
		t.Errorf("Channel not created correctly by NewChannel()")
	}
	if len(ch.nicks) != 0 || len(ch.lookup) != 0 {
		t.Errorf("Channel maps contain data after NewChannel()")
	}
}

func TestAddNick(t *testing.T) {
	_, s := setUp(t)
	defer s.tearDown()

	ch := NewChannel("#test1", s.log)
	nk := NewNick("test1", s.log)
	cp := new(ChanPrivs)

	ch.addNick(nk, cp)

	if len(ch.nicks) != 1 || len(ch.lookup) != 1 {
		t.Errorf("Nick lists not updated correctly for add.")
	}
	if c, ok := ch.nicks[nk]; !ok || c != cp {
		t.Errorf("Nick test1 not properly stored in nicks map.")
	}
	if n, ok := ch.lookup["test1"]; !ok || n != nk {
		t.Errorf("Nick test1 not properly stored in lookup map.")
	}

	s.log.EXPECT().Warn("Channel.addNick(): %s already on %s.",
		"test1", "#test1")
	ch.addNick(nk, cp)
}

func TestDelNick(t *testing.T) {
	_, s := setUp(t)
	defer s.tearDown()

	ch := NewChannel("#test1", s.log)
	nk := NewNick("test1", s.log)
	cp := new(ChanPrivs)

	// Testing the error state first is easier
	s.log.EXPECT().Warn("Channel.delNick(): %s not on %s.",
		"test1", "#test1")
	ch.delNick(nk)

	ch.addNick(nk, cp)
	ch.delNick(nk)
	if len(ch.nicks) != 0 || len(ch.lookup) != 0 {
		t.Errorf("Nick lists not updated correctly for del.")
	}
	if c, ok := ch.nicks[nk]; ok || c != nil {
		t.Errorf("Nick test1 not properly removed from nicks map.")
	}
	if n, ok := ch.lookup["#test1"]; ok || n != nil {
		t.Errorf("Nick test1 not properly removed from lookup map.")
	}
}

func TestChannelParseModes(t *testing.T) {
	_, s := setUp(t)
	defer s.tearDown()

	ch := NewChannel("#test1", s.log)
	md := ch.Modes

	// Channel modes can adjust channel privs too, so we need a Nick
	nk := NewNick("test1", s.log)
	cp := new(ChanPrivs)
	ch.addNick(nk, cp)

	// Test bools first.
	if md.Private || md.Secret || md.ProtectedTopic || md.NoExternalMsg ||
		md.Moderated || md.InviteOnly || md.OperOnly || md.SSLOnly {
		t.Errorf("Modes for new channel set to true.")
	}

	// Flip some bits!
	md.Private = true
	md.NoExternalMsg = true
	md.InviteOnly = true

	// Flip some MOAR bits.
	ch.ParseModes("+s-p+tm-i")

	if md.Private || !md.Secret || !md.ProtectedTopic || !md.NoExternalMsg ||
		!md.Moderated || md.InviteOnly || md.OperOnly || md.SSLOnly {
		t.Errorf("Modes not flipped correctly by ParseModes.")
	}

	// Test numeric parsing (currently only channel limits)
	if md.Limit != 0 {
		t.Errorf("Limit for new channel not zero.")
	}

	// enable limit correctly
	ch.ParseModes("+l", "256")
	if md.Limit != 256 {
		t.Errorf("Limit for channel not set correctly")
	}

	// enable limit incorrectly. see nick_test.go for why the byte() cast.
	s.log.EXPECT().Warn("Channel.ParseModes(): not enough arguments to "+
		"process MODE %s %s%c", "#test1", "+", byte('l'))
	ch.ParseModes("+l")
	if md.Limit != 256 {
		t.Errorf("Bad limit value caused limit to be unset.")
	}

	// disable limit correctly
	ch.ParseModes("-l")
	if md.Limit != 0 {
		t.Errorf("Limit for channel not unset correctly")
	}

	// Test string parsing (currently only channel key)
	if md.Key != "" {
		t.Errorf("Key set for new channel.")
	}

	// enable key correctly
	ch.ParseModes("+k", "foobar")
	if md.Key != "foobar" {
		t.Errorf("Key for channel not set correctly")
	}

	// enable key incorrectly
	s.log.EXPECT().Warn("Channel.ParseModes(): not enough arguments to "+
		"process MODE %s %s%c", "#test1", "+", byte('k'))
	ch.ParseModes("+k")
	if md.Key != "foobar" {
		t.Errorf("Bad key value caused key to be unset.")
	}

	// disable key correctly
	ch.ParseModes("-k")
	if md.Key != "" {
		t.Errorf("Key for channel not unset correctly")
	}

	// Test chan privs parsing.
	cp.Op = true
	cp.HalfOp = true
	ch.ParseModes("+aq-o", "test1", "test1", "test1")

	if !cp.Owner || !cp.Admin || cp.Op || !cp.HalfOp || cp.Voice {
		t.Errorf("Channel privileges not flipped correctly by ParseModes.")
	}

	s.log.EXPECT().Warn("Channel.ParseModes(): untracked nick %s "+
		"received MODE on channel %s", "test2", "#test1")
	ch.ParseModes("+v", "test2")

	s.log.EXPECT().Warn("Channel.ParseModes(): not enough arguments to "+
		"process MODE %s %s%c", "#test1", "-", byte('v'))
	ch.ParseModes("-v")

	// Test a random mix of modes, just to be sure
	md.Limit = 256
	s.log.EXPECT().Warn("Channel.ParseModes(): not enough arguments to "+
		"process MODE %s %s%c", "#test1", "-", byte('h'))
	ch.ParseModes("+zpt-qsl+kv-h", "test1", "foobar", "test1")

	if !md.Private || md.Secret || !md.ProtectedTopic || !md.NoExternalMsg ||
		!md.Moderated || md.InviteOnly || md.OperOnly || !md.SSLOnly {
		t.Errorf("Modes not flipped correctly by ParseModes (2).")
	}
	if md.Limit != 0 || md.Key != "foobar" {
		t.Errorf("Key and limit not changed correctly by ParseModes (2).")
	}
	if cp.Owner || !cp.Admin || cp.Op || !cp.HalfOp || !cp.Voice {
		// NOTE: HalfOp not actually unset above thanks to deliberate error.
		t.Errorf("Channel privileges not flipped correctly by ParseModes (2).")
	}

	// Finally, check we get an info log for an unrecognised mode character
	s.log.EXPECT().Info("Channel.ParseModes(): unknown mode char %c", byte('d'))
	ch.ParseModes("+d")
}
