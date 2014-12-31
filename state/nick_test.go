package state

import "testing"

func compareNick(t *testing.T, nk *nick) {
	n := nk.Nick()
	if n.Nick != nk.nick || n.Ident != nk.ident || n.Host != nk.host || n.Name != nk.name ||
		!n.Modes.Equals(nk.modes) || len(n.Channels) != len(nk.chans) {
		t.Errorf("Nick not duped correctly from internal state.")
	}
	for ch, cp := range nk.chans {
		if other, ok := n.Channels[ch.name]; !ok || !cp.Equals(other) {
			t.Errorf("Channel not duped correctly from internal state.")
		}
	}
}

func TestNewNick(t *testing.T) {
	nk := newNick("test1")

	if nk.nick != "test1" {
		t.Errorf("Nick not created correctly by NewNick()")
	}
	if len(nk.chans) != 0 || len(nk.lookup) != 0 {
		t.Errorf("Nick maps contain data after NewNick()")
	}
	compareNick(t, nk)
}

func TestAddChannel(t *testing.T) {
	nk := newNick("test1")
	ch := newChannel("#test1")
	cp := new(ChanPrivs)

	nk.addChannel(ch, cp)

	if len(nk.chans) != 1 || len(nk.lookup) != 1 {
		t.Errorf("Channel lists not updated correctly for add.")
	}
	if c, ok := nk.chans[ch]; !ok || c != cp {
		t.Errorf("Channel #test1 not properly stored in chans map.")
	}
	if c, ok := nk.lookup["#test1"]; !ok || c != ch {
		t.Errorf("Channel #test1 not properly stored in lookup map.")
	}
	compareNick(t, nk)
}

func TestDelChannel(t *testing.T) {
	nk := newNick("test1")
	ch := newChannel("#test1")
	cp := new(ChanPrivs)

	nk.addChannel(ch, cp)
	nk.delChannel(ch)
	if len(nk.chans) != 0 || len(nk.lookup) != 0 {
		t.Errorf("Channel lists not updated correctly for del.")
	}
	if c, ok := nk.chans[ch]; ok || c != nil {
		t.Errorf("Channel #test1 not properly removed from chans map.")
	}
	if c, ok := nk.lookup["#test1"]; ok || c != nil {
		t.Errorf("Channel #test1 not properly removed from lookup map.")
	}
	compareNick(t, nk)
}

func TestNickParseModes(t *testing.T) {
	nk := newNick("test1")
	md := nk.modes

	// Modes should all be false for a new nick
	if md.Invisible || md.Oper || md.WallOps || md.HiddenHost || md.SSL {
		t.Errorf("Modes for new nick set to true.")
	}

	// Set a couple of modes, for testing.
	md.Invisible = true
	md.HiddenHost = true

	// Parse a mode line that flips one true to false and two false to true
	nk.parseModes("+z-x+w")

	compareNick(t, nk)
	if !md.Invisible || md.Oper || !md.WallOps || md.HiddenHost || !md.SSL {
		t.Errorf("Modes not flipped correctly by ParseModes.")
	}
}
