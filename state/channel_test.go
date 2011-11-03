package state

import (
	"github.com/fluffle/goirc/logging"
	"testing"
)

func TestNewChannel(t *testing.T) {
	l, _ := logging.NewMock()
	ch := NewChannel("#test1", l)

	if ch.Name != "#test1" || ch.l != l {
		t.Errorf("Channel not created correctly by NewChannel()")
	}
	if len(ch.nicks) != 0 || len(ch.lookup) != 0 {
		t.Errorf("Channel maps contain data after NewChannel()")
	}
}

func TestAddNick(t *testing.T) {
	l, m := logging.NewMock()
	ch := NewChannel("#test1", l)
	nk := NewNick("test1", l)
	cp := new(ChanPrivs)

	ch.addNick(nk, cp)
	m.CheckNothingWritten(t)

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
	m.CheckWrittenAtLevel(t, logging.Warn,
		"Channel.addNick(): test1 already on #test1.")
}

func TestDelNick(t *testing.T) {
	l, m := logging.NewMock()
	ch := NewChannel("#test1", l)
	nk := NewNick("test1", l)
	cp := new(ChanPrivs)

	// Testing the error state first is easier
	ch.delNick(nk)
	m.CheckWrittenAtLevel(t, logging.Warn,
		"Channel.delNick(): test1 not on #test1.")

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
