package logging

import (
	"os"
	"strings"
	"testing"
)

// These are provided to ease testing code that uses the logging pkg

type mockWriter struct {
	written []byte
}

func (w *mockWriter) Write(p []byte) (n int, err os.Error) {
	w.written = append(w.written, p...)
	return len(p), nil
}

func (w *mockWriter) reset() {
	w.written = w.written[:0]
}

type writerMap map[LogLevel]*mockWriter

// This doesn't create a mock Logger but a Logger that writes to mock outputs
func NewMock() (*logger, writerMap) {
	wMap := writerMap{
		Debug: &mockWriter{make([]byte, 0)},
		Info:  &mockWriter{make([]byte, 0)},
		Warn:  &mockWriter{make([]byte, 0)},
		Error: &mockWriter{make([]byte, 0)},
		Fatal: &mockWriter{make([]byte, 0)},
	}
	logMap := make(LogMap)
	for lv, w := range wMap {
		logMap[lv] = makeLogger(w)
	}
	return New(logMap, Error, false), wMap
}

func (m writerMap) CheckNothingWritten(t *testing.T) {
	for lv, w := range m {
		if len(w.written) > 0 {
			t.Errorf("%d bytes logged at level %s, expected none:",
				len(w.written), logString[lv])
			t.Errorf("\t%s", w.written)
			w.reset()
		}
	}
}

func (m writerMap) CheckWrittenAtLevel(t *testing.T, lv LogLevel, exp string) {
	var w *mockWriter
	if _, ok := m[lv]; !ok {
		w = m[Debug]
	} else {
		w = m[lv]
	}
	// 32 bytes covers the date, time and filename up to the colon in
	// 2011/10/22 10:22:57 log_test.go:<line no>: <level> <log message>
	if len(w.written) <= 33 {
		t.Errorf("Not enough bytes logged at level %s:", LogString(lv))
		t.Errorf("\tgot: %s", string(w.written))
		return
	}
	s := string(w.written[32:])
	// 2 covers the : itself and the extra space
	idx := strings.Index(s, ":") + 2
	// s will end in "\n", so -1 to chop that off too
	s = s[idx:len(s)-1]
	// expected won't contain the log level prefix, so prepend that
	exp = LogString(lv) + " " + exp
	if s != exp {
		t.Errorf("Log message at level %s differed.", LogString(lv))
		t.Errorf("\texp: %s\n\tgot: %s", exp, s)
	}
	w.reset()
	// Calling checkNothingWritten here both tests that w.reset() works
	// and verifies that nothing was written at any other levels.
	m.CheckNothingWritten(t)
}

