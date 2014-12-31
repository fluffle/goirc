package state

import "testing"

func compareChannel(t *testing.T, ch *channel) {
	c := ch.Channel()
	if c.Name != ch.name || c.Topic != ch.topic ||
		!c.Modes.Equals(ch.modes) || len(c.Nicks) != len(ch.nicks) {
		t.Errorf("Channel not duped correctly from internal state.")
	}
	for nk, cp := range ch.nicks {
		if other, ok := c.Nicks[nk.nick]; !ok || !cp.Equals(other) {
			t.Errorf("Nick not duped correctly from internal state.")
		}
	}
}

func TestNewChannel(t *testing.T) {
	ch := newChannel("#test1")

	if ch.name != "#test1" {
		t.Errorf("Channel not created correctly by NewChannel()")
	}
	if len(ch.nicks) != 0 || len(ch.lookup) != 0 {
		t.Errorf("Channel maps contain data after NewChannel()")
	}
	compareChannel(t, ch)
}

func TestAddNick(t *testing.T) {
	ch := newChannel("#test1")
	nk := newNick("test1")
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
	compareChannel(t, ch)
}

func TestDelNick(t *testing.T) {
	ch := newChannel("#test1")
	nk := newNick("test1")
	cp := new(ChanPrivs)

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
	compareChannel(t, ch)
}

func TestChannelParseModes(t *testing.T) {
	ch := newChannel("#test1")
	md := ch.modes

	// Channel modes can adjust channel privs too, so we need a Nick
	nk := newNick("test1")
	cp := new(ChanPrivs)
	ch.addNick(nk, cp)

	// Test bools first.
	compareChannel(t, ch)
	if md.Private || md.Secret || md.ProtectedTopic || md.NoExternalMsg ||
		md.Moderated || md.InviteOnly || md.OperOnly || md.SSLOnly {
		t.Errorf("Modes for new channel set to true.")
	}

	// Flip some bits!
	md.Private = true
	md.NoExternalMsg = true
	md.InviteOnly = true

	// Flip some MOAR bits.
	ch.parseModes("+s-p+tm-i")

	compareChannel(t, ch)
	if md.Private || !md.Secret || !md.ProtectedTopic || !md.NoExternalMsg ||
		!md.Moderated || md.InviteOnly || md.OperOnly || md.SSLOnly {
		t.Errorf("Modes not flipped correctly by ParseModes.")
	}

	// Test numeric parsing (currently only channel limits)
	if md.Limit != 0 {
		t.Errorf("Limit for new channel not zero.")
	}

	// enable limit correctly
	ch.parseModes("+l", "256")
	compareChannel(t, ch)
	if md.Limit != 256 {
		t.Errorf("Limit for channel not set correctly")
	}

	// enable limit incorrectly
	ch.parseModes("+l")
	compareChannel(t, ch)
	if md.Limit != 256 {
		t.Errorf("Bad limit value caused limit to be unset.")
	}

	// disable limit correctly
	ch.parseModes("-l")
	compareChannel(t, ch)
	if md.Limit != 0 {
		t.Errorf("Limit for channel not unset correctly")
	}

	// Test string parsing (currently only channel key)
	if md.Key != "" {
		t.Errorf("Key set for new channel.")
	}

	// enable key correctly
	ch.parseModes("+k", "foobar")
	compareChannel(t, ch)
	if md.Key != "foobar" {
		t.Errorf("Key for channel not set correctly")
	}

	// enable key incorrectly
	ch.parseModes("+k")
	compareChannel(t, ch)
	if md.Key != "foobar" {
		t.Errorf("Bad key value caused key to be unset.")
	}

	// disable key correctly
	ch.parseModes("-k")
	compareChannel(t, ch)
	if md.Key != "" {
		t.Errorf("Key for channel not unset correctly")
	}

	// Test chan privs parsing.
	cp.Op = true
	cp.HalfOp = true
	ch.parseModes("+aq-o", "test1", "test1", "test1")

	compareChannel(t, ch)
	if !cp.Owner || !cp.Admin || cp.Op || !cp.HalfOp || cp.Voice {
		t.Errorf("Channel privileges not flipped correctly by ParseModes.")
	}

	// Test a random mix of modes, just to be sure
	md.Limit = 256
	ch.parseModes("+zpt-qsl+kv-h", "test1", "foobar", "test1")

	compareChannel(t, ch)
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
}
