package logging

import (
	"os"
	"testing"
)

type mockWriter struct {
	written []byte
}

func (w *mockWriter) Write(p []byte) (n int, err os.Error) {
	w.written = append(w.written, p...)
	return len(p), nil
}

func TestDefaultLogging(t *testing.T) {
	w := &mockWriter{make([]byte, 0)}
	l := New(w, LogError, false)
	l.Log(4, "Nothing should be logged yet")
	l.Log(LogDebug, "or yet...")
	l.Log(LogWarn, "or yet!")
	if len(w.written) > 0 {
		t.Errorf("Unexpected low-level logging output.")
	}
	l.Log(LogError, "Error!")
	// Note: the below is deliberately fragile to ensure
	// the right file:line is logged on errors. Sorry!
	if s := string(w.written); s[20:] != "log_test.go:26: Error!\n" {
		t.Errorf("Error incorrectly logged (check line numbers!)")
	}
}
