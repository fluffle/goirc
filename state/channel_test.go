package state

import (
	"github.com/fluffle/goirc/logging"
	"testing"
)

func TestNewChannel(t *testing.T) {
	l, _ := logging.NewMock(t)
	ch := NewChannel("#test1", l)

	if ch.Name != "#test1" || ch.l != l {
		t.Errorf("Channel not created correctly by NewChannel()")
	}
	if len(ch.nicks) != 0 || len(ch.lookup) != 0 {
		t.Errorf("Channel maps contain data after NewChannel()")
	}
}

func TestAddNick(t *testing.T) {
	l, m := logging.NewMock(t)
	ch := NewChannel("#test1", l)
	nk := NewNick("test1", l)
	cp := new(ChanPrivs)

	ch.addNick(nk, cp)
	m.ExpectNothing()

	if len(ch.nicks) != 1 || len(ch.lookup) != 1 {
		t.Errorf("Nick lists not updated correctly for add.")
	}
	if c, ok := ch.nicks[nk]; !ok || c != cp {
		t.Errorf("Nick test1 not properly stored in nicks map.")
	}
	if n, ok := ch.lookup["test1"]; !ok || n != nk {
		t.Errorf("Nick test1 not properly stored in lookup map.")
	}

	ch.addNick(nk, cp)
	m.ExpectAt(logging.Warn, "Channel.addNick(): test1 already on #test1.")
}

func TestDelNick(t *testing.T) {
	l, m := logging.NewMock(t)
	ch := NewChannel("#test1", l)
	nk := NewNick("test1", l)
	cp := new(ChanPrivs)

	// Testing the error state first is easier
	ch.delNick(nk)
	m.ExpectAt(logging.Warn, "Channel.delNick(): test1 not on #test1.")

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
	l, m := logging.NewMock(t)
	ch := NewChannel("#test1", l)
	md := ch.Modes

	// Channel modes can adjust channel privs too, so we need a Nick
	nk := NewNick("test1", l)
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
	m.ExpectNothing()

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
	m.ExpectNothing()
	if md.Limit != 256 {
		t.Errorf("Limit for channel not set correctly")
	}

	// enable limit incorrectly
	ch.ParseModes("+l")
	m.ExpectAt(logging.Warn,
		"Channel.ParseModes(): not enough arguments to process MODE #test1 +l")
	if md.Limit != 256 {
		t.Errorf("Bad limit value caused limit to be unset.")
	}

	// disable limit correctly
	ch.ParseModes("-l")
	m.ExpectNothing()
	if md.Limit != 0 {
		t.Errorf("Limit for channel not unset correctly")
	}

	// Test string parsing (currently only channel key)
	if md.Key != "" {
		t.Errorf("Key set for new channel.")
	}

	// enable key correctly
	ch.ParseModes("+k", "foobar")
	m.ExpectNothing()
	if md.Key != "foobar" {
		t.Errorf("Key for channel not set correctly")
	}

	// enable key incorrectly
	ch.ParseModes("+k")
	m.ExpectAt(logging.Warn,
		"Channel.ParseModes(): not enough arguments to process MODE #test1 +k")
	if md.Key != "foobar" {
		t.Errorf("Bad key value caused key to be unset.")
	}

	// disable key correctly
	ch.ParseModes("-k")
	m.ExpectNothing()
	if md.Key != "" {
		t.Errorf("Key for channel not unset correctly")
	}

	// Test chan privs parsing.
	cp.Op = true
	cp.HalfOp = true
	ch.ParseModes("+aq-o", "test1", "test1", "test1")
	m.ExpectNothing()

	if !cp.Owner || !cp.Admin || cp.Op || !cp.HalfOp || cp.Voice {
		t.Errorf("Channel privileges not flipped correctly by ParseModes.")
	}

	ch.ParseModes("+v", "test2")
	m.ExpectAt(logging.Warn,
		"Channel.ParseModes(): untracked nick test2 received MODE on channel #test1")

	ch.ParseModes("-v")
	m.ExpectAt(logging.Warn,
		"Channel.ParseModes(): not enough arguments to process MODE #test1 -v")
	
	// Test a random mix of modes, just to be sure
	md.Limit = 256
	ch.ParseModes("+zpt-qsl+kv-h", "test1", "foobar", "test1")
	m.ExpectAt(logging.Warn,
		"Channel.ParseModes(): not enough arguments to process MODE #test1 -h")

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
	ch.ParseModes("+d")
	m.ExpectAt(logging.Info, "Channel.ParseModes(): unknown mode char d")
}
