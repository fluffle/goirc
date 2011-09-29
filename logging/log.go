package logging

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"sync"
)

// A simple level-based logging system.
// Note that higher levels of logging are still usable via Log()
// Also, remember to call flag.Parse() near the start of your func main()
const (
	LogFatal = iota - 1
	LogError
	LogWarn
	LogInfo
	LogDebug
)

var (
	file = flag.String("log.file", "",
		"Log to this file rather than STDERR")
	level = flag.Int("log.level", LogError,
		"Level of logging to be output")
	only = flag.Bool("log.only", false,
		"Only log output at the selected level")

	// Shortcut flags for great justice
	quiet = flag.Bool("log.quiet", false,
		"Only fatal output (equivalent to -v -1)")
	warn = flag.Bool("log.warn", false,
		"Warning output (equivalent to -v 1)")
	info = flag.Bool("log.info", false,
		"Info output (equivalent to -v 2)")
	debug = flag.Bool("log.debug", false,
		"Debug output (equivalent to -v 3)")
)

type Logger interface {
	// Log at a given level
	Log(int, string, ...interface{})
	// Log at level 3
	Debug(string, ...interface{})
	// Log at level 2
	Info(string, ...interface{})
	// Log at level 1
	Warn(string, ...interface{})
	// Log at level 0
	Error(string, ...interface{})
	// Log at level -1, to STDERR always, and exit after logging.
	Fatal(string, ...interface{})
	// Change the current log display level
	SetLogLevel(int)
}

// A struct to implement the above interface
type logger struct {
	// We wrap a log.Logger for most of the heavy lifting
	// but it can't be anonymous thanks to the conflicting definitions of Fatal
	log         *log.Logger
	level       int
	only        bool
	*sync.Mutex // to ensure changing levels/flags is atomic
}

var internal *logger

func init() {
	/// Hopefully this won't cause pain and suffering
	flag.Parse()

	// Where are we logging to?
	var out io.Writer
	if *file != "" {
		fh, err := os.OpenFile(*file, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
		if err != nil {
			log.Fatalf("Error opening log file: %s", err)
		} else {
			out = fh
		}
	} else {
		out = os.Stderr
	}

	// What are we logging?
	var lv int
	// The shortcut flags prioritize by level, but an
	// explicit level flag takes first precedence.
	// I think the switch looks cleaner than if/else if, meh :-)
	switch {
	case *level != 0:
		lv = *level
	case *quiet:
		lv = LogFatal
	case *warn:
		lv = LogWarn
	case *info:
		lv = LogInfo
	case *debug:
		lv = LogDebug
	}

	internal = New(out, lv, *only)
}

func New(out io.Writer, level int, only bool) *logger {
	l := log.New(out, "", log.LstdFlags)
	return &logger{l, level, only, &sync.Mutex{}}
}

// Internal function all others call to ensure identical call depth
func (l *logger) write(lv int, fm string, v ...interface{}) {
	if lv > l.level {
		// Your logs are not important to us, goodnight
		return
	}

	l.Lock()
	defer l.Unlock()
	lineno := bool((l.log.Flags() & log.Lshortfile) > 0)
	// Enable logging file:line if LogWarn level or worse
	if lv <= LogWarn && !lineno {
		l.log.SetFlags(log.LstdFlags | log.Lshortfile)
	} else if lv > LogWarn && lineno {
		l.log.SetFlags(log.LstdFlags)
	}
	// Writing the log is deceptively simple
	l.log.Output(3, fmt.Sprintf(fm, v...))
	if lv == LogFatal {
		// Always fatal to stderr too.
		log.Fatalf(fm, v...)
	}
}

func (l *logger) Log(lv int, fm string, v ...interface{}) {
	l.write(lv, fm, v...)
}

func Log(lv int, fm string, v ...interface{}) {
	internal.write(lv, fm, v...)
}

// Helper functions for specific levels
func (l *logger) Debug(fm string, v ...interface{}) {
	l.write(LogDebug, fm, v...)
}

func Debug(fm string, v ...interface{}) {
	internal.write(LogDebug, fm, v...)
}

func (l *logger) Info(fm string, v ...interface{}) {
	l.write(LogInfo, fm, v...)
}

func Info(fm string, v ...interface{}) {
	internal.write(LogInfo, fm, v...)
}

func (l *logger) Warn(fm string, v ...interface{}) {
	l.write(LogWarn, fm, v...)
}

func Warn(fm string, v ...interface{}) {
	internal.write(LogWarn, fm, v...)
}

func (l *logger) Error(fm string, v ...interface{}) {
	l.write(LogError, fm, v...)
}

func Error(fm string, v ...interface{}) {
	internal.write(LogError, fm, v...)
}

func (l *logger) Fatal(fm string, v ...interface{}) {
	l.write(LogFatal, fm, v...)
}

func Fatal(fm string, v ...interface{}) {
	internal.write(LogFatal, fm, v...)
}

func (l *logger) SetLogLevel(lv int) {
	l.Lock()
	defer l.Unlock()
	l.level = lv
}

func SetLogLevel(lv int) {
	internal.SetLogLevel(lv)
}
