package logging

import (
	"testing"
)

// Note: the below is deliberately PLACED AT THE TOP OF THIS FILE because
// it is fragile. It ensures the right file:line is logged. Sorry!
func TestLogCorrectLineNumbers(t *testing.T) {
	l, m := NewMock(t)
	l.Log(Error, "Error!")
	// This breaks the mock encapsulation a little, but meh.
	if s := string(m.m[Error].written); s[20:] != "log_test.go:11: ERROR Error!\n" {
		t.Errorf("Error incorrectly logged (check line numbers!)")
	}
}

func TestStandardLogging(t *testing.T) {
	l, m := NewMock(t)
	l.SetLogLevel(Error)

	l.Log(4, "Nothing should be logged yet")
	m.ExpectNothing()

	l.Log(Debug, "or yet...")
	m.ExpectNothing()

	l.Log(Info, "or yet...")
	m.ExpectNothing()

	l.Log(Warn, "or yet!")
	m.ExpectNothing()

	l.Log(Error, "Error!")
	m.Expect("Error!")
}

func TestAllLoggingLevels(t *testing.T) {
	l, m := NewMock(t)

	l.Log(4, "Log to level 4.")
	m.ExpectAt(4, "Log to level 4.")

	l.Debug("Log to debug.")
	m.ExpectAt(Debug, "Log to debug.")

	l.Info("Log to info.")
	m.ExpectAt(Info, "Log to info.")

	l.Warn("Log to warning.")
	m.ExpectAt(Warn, "Log to warning.")

	l.Error("Log to error.")
	m.ExpectAt(Error, "Log to error.")

	// recover to track the panic caused by Fatal.
	defer func() { recover() }()
	l.Fatal("Log to fatal.")
	m.ExpectAt(Fatal, "Log to fatal.")
}
