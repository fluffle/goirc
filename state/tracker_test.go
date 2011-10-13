package state

import (
	"testing"
)

func TestNewNick(t *testing.T) {
	st := NewTracker()

	if len(st.nicks) != 0 {
		t.Errorf("Nick list of new tracker is non-zero length.")
	}

	nick := st.NewNick("test1")

	if nick == nil || nick.Nick != "test1" || nick.st != st {
		t.Errorf("Nick object created incorrectly by NewNick.")
	}
	if n, ok := st.nicks["test1"]; !ok || n != nick || len(st.nicks) != 1 {
		t.Errorf("Nick object stored incorrectly by NewNick.")
	}

	if fail := st.NewNick("test1"); fail != nil {
		t.Errorf("Creating duplicate nick did not produce nil return.")
	}
}

func TestGetNick(t *testing.T) {
	st := NewTracker()

	test1 := &Nick{Nick: "test1", st: st}
	st.nicks["test1"] = test1

	if n := st.GetNick("test1"); n != test1 {
		t.Errorf("Incorrect nick returned by GetNick.")
	}
	if n := st.GetNick("test2"); n != nil {
		t.Errorf("Nick unexpectedly returned by GetNick.")
	}
	if len(st.nicks) != 1 {
		t.Errorf("Nick list changed size during GetNick.")
	}
}

func TestReNick(t *testing.T) {
	st := NewTracker()

	test1 := &Nick{Nick: "test1", st: st}
	st.nicks["test1"] = test1

	st.ReNick("test1", "test2")

	if _, ok := st.nicks["test1"]; ok {
		t.Errorf("Nick test1 still exists after ReNick.")
	}
	if n, ok := st.nicks["test2"]; !ok || n != test1 {
		t.Errorf("Nick test2 doesn't exist after ReNick.")
	}
	if len(st.nicks) != 1 {
		t.Errorf("Nick list changed size during ReNick.")
	}

	test2 := &Nick{Nick: "test2", st: st}
	st.nicks["test1"] = test2

	st.ReNick("test1", "test2")
	if n, ok := st.nicks["test2"]; !ok || n != test1 {
		t.Errorf("Nick test2 overwritten/deleted by ReNick.")
	}
	if n, ok := st.nicks["test1"]; !ok || n != test2 {
		t.Errorf("Nick test1 overwritten/deleted by ReNick.")
	}
	if len(st.nicks) != 2 {
		t.Errorf("Nick list changed size during ReNick.")
	}

}

func TestDelNick(t *testing.T) {
	st := NewTracker()

	test1 := &Nick{Nick: "test1", st: st}
	st.nicks["test1"] = test1

	st.DelNick("test1")

	if _, ok := st.nicks["test1"]; ok {
		t.Errorf("Nick test1 still exists after DelNick.")
	}
	if len(st.nicks) != 0 {
		t.Errorf("Nick list still contains nicks after DelNick.")
	}

	st.nicks["test1"] = test1

	st.DelNick("test2")

	if len(st.nicks) != 1 {
		t.Errorf("DelNick had unexpected side-effects.")
	}
}
