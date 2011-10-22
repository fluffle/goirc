package state

import (
	"github.com/fluffle/goirc/logging"
	"testing"
)

func TestNewNick(t *testing.T) {
	l, _ := logging.NewMock()
	nk := NewNick("test1", l)

	if nk.Nick != "test1" || nk.l != l {
		t.Errorf("Nick not created correctly by NewNick()")
	}
	if len(nk.chans) != 0 || len(nk.lookup) != 0 {
		t.Errorf("Nick maps contain data after NewNick()")
	}
}

func TestAddChannel(t *testing.T) {
	l, m := logging.NewMock()
	nk := NewNick("test1", l)
	ch := NewChannel("#test1", l)
	cp := new(ChanPrivs)

	nk.addChannel(ch, cp)
	m.CheckNothingWritten(t)

	if len(nk.chans) != 1 || len(nk.lookup) != 1 {
		t.Errorf("Channel lists not updated correctly.")
	}
	if c, ok := nk.chans[ch]; !ok || c != cp {
		t.Errorf("Channel #test1 not properly stored in chans map.")
	}
	if c, ok := nk.lookup["#test1"]; !ok || c != ch {
		t.Errorf("Channel #test1 not properly stored in lookup map.")
	}
}
