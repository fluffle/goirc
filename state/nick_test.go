package state

import (
	"github.com/fluffle/goirc/logging"
	"testing"
)

func TestNewNick(t *testing.T) {
	l, _ := logging.NewMock(t)
	nk := NewNick("test1", l)

	if nk.Nick != "test1" || nk.l != l {
		t.Errorf("Nick not created correctly by NewNick()")
	}
	if len(nk.chans) != 0 || len(nk.lookup) != 0 {
		t.Errorf("Nick maps contain data after NewNick()")
	}
}

func TestAddChannel(t *testing.T) {
	l, m := logging.NewMock(t)
	nk := NewNick("test1", l)
	ch := NewChannel("#test1", l)
	cp := new(ChanPrivs)

	nk.addChannel(ch, cp)
	m.ExpectNothing()

	if len(nk.chans) != 1 || len(nk.lookup) != 1 {
		t.Errorf("Channel lists not updated correctly for add.")
	}
	if c, ok := nk.chans[ch]; !ok || c != cp {
		t.Errorf("Channel #test1 not properly stored in chans map.")
	}
	if c, ok := nk.lookup["#test1"]; !ok || c != ch {
		t.Errorf("Channel #test1 not properly stored in lookup map.")
	}

	nk.addChannel(ch, cp)
	m.ExpectAt(logging.Warn, "Nick.addChannel(): test1 already on #test1.")
}

func TestDelChannel(t *testing.T) {
	l, m := logging.NewMock(t)
	nk := NewNick("test1", l)
	ch := NewChannel("#test1", l)
	cp := new(ChanPrivs)

	// Testing the error state first is easier
	nk.delChannel(ch)
	m.ExpectAt(logging.Warn, "Nick.delChannel(): test1 not on #test1.")

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
}

func TestNickParseModes(t *testing.T) {
	l, m := logging.NewMock(t)
	nk := NewNick("test1", l)
	md := nk.Modes

	// Modes should all be false for a new nick
	if md.Invisible || md.Oper || md.WallOps || md.HiddenHost || md.SSL {
		t.Errorf("Modes for new nick set to true.")
	}

	// Set a couple of modes, for testing.
	md.Invisible = true
	md.HiddenHost = true

	// Parse a mode line that flips one true to false and two false to true
	nk.ParseModes("+z-x+w")
	m.ExpectNothing()

	if !md.Invisible || md.Oper || !md.WallOps || md.HiddenHost || !md.SSL {
		t.Errorf("Modes not flipped correctly by ParseModes.")
	}

	// Check that passing an unknown mode char results in an info log
	nk.ParseModes("+d")
	m.ExpectAt(logging.Info, "Nick.ParseModes(): unknown mode char d")
}
