package logging

import (
	"testing"
)

// Note: the below is deliberately PLACED AT THE TOP OF THIS FILE because
// it is fragile. It ensures the right file:line is logged. Sorry!
func TestLogCorrectLineNumbers(t *testing.T) {
	l, m := NewMock()
	l.Log(Error, "Error!")
	if s := string(m[Error].written); s[20:] != "log_test.go:11: ERROR Error!\n" {
		t.Errorf("Error incorrectly logged (check line numbers!)")
	}
}

func TestStandardLogging(t *testing.T) {
	l, m := NewMock()
	l.SetLogLevel(Error)

	l.Log(4, "Nothing should be logged yet")
	m.CheckNothingWritten(t)

	l.Log(Debug, "or yet...")
	m.CheckNothingWritten(t)

	l.Log(Info, "or yet...")
	m.CheckNothingWritten(t)

	l.Log(Warn, "or yet!")
	m.CheckNothingWritten(t)

	l.Log(Error, "Error!")
	m.CheckWrittenAtLevel(t, Error, "Error!")
}

func TestAllLoggingLevels(t *testing.T) {
	l, m := NewMock()
	l.SetLogLevel(5)

	l.Log(4, "Log to level 4.")
	m.CheckWrittenAtLevel(t, 4, "Log to level 4.")

	l.Debug("Log to debug.")
	m.CheckWrittenAtLevel(t, Debug, "Log to debug.")

	l.Info("Log to info.")
	m.CheckWrittenAtLevel(t, Info, "Log to info.")

	l.Warn("Log to warning.")
	m.CheckWrittenAtLevel(t, Warn, "Log to warning.")

	l.Error("Log to error.")
	m.CheckWrittenAtLevel(t, Error, "Log to error.")

	// recover to track the panic caused by Fatal.
	defer func() { recover() }()
	l.Fatal("Log to fatal.")
	m.CheckWrittenAtLevel(t, Fatal, "Log to fatal.")
}
