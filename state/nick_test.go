package state

import (
	"testing"
)

func TestNewNick(t *testing.T) {
	_, s := setUp(t)
	defer s.tearDown()

	nk := NewNick("test1", s.log)

	if nk.Nick != "test1" || nk.l != s.log {
		t.Errorf("Nick not created correctly by NewNick()")
	}
	if len(nk.chans) != 0 || len(nk.lookup) != 0 {
		t.Errorf("Nick maps contain data after NewNick()")
	}
}

func TestAddChannel(t *testing.T) {
	_, s := setUp(t)
	defer s.tearDown()

	nk := NewNick("test1", s.log)
	ch := NewChannel("#test1", s.log)
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

	s.log.EXPECT().Warn("Nick.addChannel(): %s already on %s.",
		"test1", "#test1")
	nk.addChannel(ch, cp)
}

func TestDelChannel(t *testing.T) {
	_, s := setUp(t)
	defer s.tearDown()

	nk := NewNick("test1", s.log)
	ch := NewChannel("#test1", s.log)
	cp := new(ChanPrivs)

	// Testing the error state first is easier
	s.log.EXPECT().Warn("Nick.delChannel(): %s not on %s.", "test1", "#test1")
	nk.delChannel(ch)

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
	_, s := setUp(t)
	defer s.tearDown()

	nk := NewNick("test1", s.log)
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

	if !md.Invisible || md.Oper || !md.WallOps || md.HiddenHost || !md.SSL {
		t.Errorf("Modes not flipped correctly by ParseModes.")
	}

	// Check that passing an unknown mode char results in an info log
	// The cast to byte here is needed to pass; gomock uses reflect.DeepEqual
	// to examine argument equality, but 'd' (when not implicitly coerced to a
	// uint8 by the type system) is an int, whereas string("+d")[1] is not.
	// This type difference (despite the values being nominally the same)
	// causes the test to fail with the following confusing error.
	//
	// no matching expected call: *logging.MockLogger.Info([Nick.ParseModes(): unknown mode char %c [100]])
	// missing call(s) to *logging.MockLogger.Info(is equal to Nick.ParseModes(): unknown mode char %c, is equal to [100])

	s.log.EXPECT().Info("Nick.ParseModes(): unknown mode char %c", byte('d'))
	nk.ParseModes("+d")
}
