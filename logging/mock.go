package logging

import (
	"os"
	"strings"
	"testing"
)

// These are provided to ease testing code that uses the logging pkg

// TODO(fluffle): Assumes at most one logging line will be written
// between calls to Expect*. Change to be Expect(exp []string)?

type mockWriter struct {
	written []byte
}

func (w *mockWriter) Write(p []byte) (n int, err os.Error) {
	w.written = append(w.written, p...)
	return len(p), nil
}

func (w *mockWriter) getLine() string {
	// 20 bytes covers the date and time in
	// 2011/10/22 10:22:57 <file>:<line>: <level> <log message>
	if len(w.written) < 20 {
		return ""
	}
	s := string(w.written)
	idx := strings.Index(s, "\n")
	s = s[20:idx]
	w.written = w.written[idx+1:]
	// consume '<file>:<line>: '
	idx = strings.Index(s, ":") + 1
	idx += strings.Index(s[idx:], ":") + 2
	return s[idx:]
}

func (w *mockWriter) reset() {
	w.written = w.written[:0]
}

type WriterMap struct {
	t *testing.T
	m map[LogLevel]*mockWriter
}

// This doesn't create a mock Logger but a Logger that writes to mock outputs
func NewMock(t *testing.T) (*logger, *WriterMap) {
	wMap := &WriterMap{
		t: t,
		m: map[LogLevel]*mockWriter{
			Debug: &mockWriter{make([]byte, 0)},
			Info:  &mockWriter{make([]byte, 0)},
			Warn:  &mockWriter{make([]byte, 0)},
			Error: &mockWriter{make([]byte, 0)},
			Fatal: &mockWriter{make([]byte, 0)},
		},
	}
	logMap := make(LogMap)
	for lv, w := range wMap.m {
		logMap[lv] = makeLogger(w)
	}
	// Set the default log level high enough that everything will get logged
	return New(logMap, (1 << 31) - 1, false), wMap
}

// When you expect something to be logged but don't care so much what level at.
func (wm *WriterMap) Expect(exp string) {
	found := false
	for lv, w := range wm.m {
		if s := w.getLine(); s != "" && !found {
			// Since we don't know what log level we're expecting, compare
			// exp against the log line with the level stripped.
			idx := strings.Index(s, " ") + 1
			if s[idx:] == exp {
				found = true
			} else {
				wm.t.Errorf("Unexpected log message encountered at level %s:",
					LogString(lv))
				wm.t.Errorf("exp: %s\ngot: %s", exp, s[idx:])
			}
		}
	}
	wm.ExpectNothing()
	if !found {
		wm.t.Errorf("Expected log message not encountered:")
		wm.t.Errorf("exp: %s", exp)
	}
}


// When you expect nothing to be logged
func (wm *WriterMap) ExpectNothing() {
	for lv, w := range wm.m {
		if s := w.getLine(); s != "" {
			wm.t.Errorf("Unexpected log message at level %s:",
				LogString(lv))
			wm.t.Errorf("%s", s)
			w.reset()
		}
	}
}

// When you expect something to be logged at a specific level.
func (wm *WriterMap) ExpectAt(lv LogLevel, exp string) {
	var w *mockWriter
	if _, ok := wm.m[lv]; !ok {
		w = wm.m[Debug]
	} else {
		w = wm.m[lv]
	}
	s := w.getLine()
	exp = strings.Join([]string{LogString(lv), exp}, " ")
	if s == "" {
		wm.t.Errorf("Nothing logged at level %s:", LogString(lv))
		wm.t.Errorf("exp: %s", exp)
		// Check nothing was written to a different log level here, too.
		wm.ExpectNothing()
		return
	}
	if s != exp {
		wm.t.Errorf("Log message at level %s differed.", LogString(lv))
		wm.t.Errorf("exp: %s\ngot: %s", exp, s)
	}
	wm.ExpectNothing()
}
