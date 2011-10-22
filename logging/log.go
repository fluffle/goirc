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

// Note that higher levels of logging are still usable via Log(). They will be
// output to the debug log in split mode if --log.level is set high enough.

// Also, remember to call flag.Parse() near the start of your func main()!

// The enforced singleton style of the standard "log" pkg is very nice, but
// it encourages people to write less testable code, and while logging is one
// of the few places where a singleton is not necessarily bad practise, it's
// not *that* hard to propagate your logging to where it needs to be.
// Alternatively you can create your own damn singleton with this package ;-)

type LogLevel int
type LogMap map[LogLevel]*log.Logger

const (
	Fatal LogLevel = iota - 1
	Error
	Warn
	Info
	Debug
)

var logString map[LogLevel]string = map[LogLevel]string{
	Fatal: "FATAL",
	Error: "ERROR",
	Warn:  "WARN",
	Info:  "INFO",
	Debug: "DEBUG",
}
func LogString(lv LogLevel) string {
	if s, ok := logString[lv]; ok {
		return s
	}
	return fmt.Sprintf("LOG(%d)", lv)
}

var (
	file = flag.String("log.file", "",
		"Log to this file rather than STDERR")
	level = flag.Int("log.level", int(Error),
		"Level of logging to be output")
	only = flag.Bool("log.only", false,
		"Only log output at the selected level")
	split = flag.Bool("log.split", false,
		"Log to one file per log level Error/Warn/Info/Debug.")

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
	SetLogLevel(LogLevel)
	// Set the logger to only output the current level
	SetOnly(bool)
}

// A struct to implement the above interface
type logger struct {
	// We wrap a set of log.Logger for most of the heavy lifting
	// but it can't be anonymous thanks to the conflicting definitions of Fatal
	log         LogMap
	level       LogLevel
	only        bool
	*sync.Mutex // to ensure changing levels/flags is atomic
}

// Helper function for opening log files, causes lots of Fatal :-)
func openLog(fn string) *log.Logger {
	fh, err := os.OpenFile(fn, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		log.Fatalf("Error opening log file: %s", err)
	}
	return makeLogger(fh)
}

// Helper function to create log.Loggers out of io.Writers
func makeLogger(w io.Writer) *log.Logger {
	return log.New(w, "", log.LstdFlags | log.Lshortfile)
}

// Creates a new logger object using the flags declared above.
// You MUST call flag.Parse before calling this ;-)
// Calling this more than once is inadvisable, you may get log corruption.
func NewFromFlags() *logger {
	// Sanity checks: if log.split is set, must have a log.file.
	if *split && *file == "" {
		log.Fatalf("You must pass --log.file with --log.split")
	}

	lv := Error
	logMap := make(LogMap)

	// What are we logging?
	// The shortcut flags prioritize by level, but an
	// explicit level flag takes first precedence.
	// I think the switch looks cleaner than if/else if, meh :-)
	switch {
	case *level != 0:
		lv = LogLevel(*level)
	case *quiet:
		lv = Fatal
	case *warn:
		lv = Warn
	case *info:
		lv = Info
	case *debug:
		lv = Debug
	}

	// Where are we logging to?
	if *split {
		// Fill in the logger map.
		for l := Fatal; l <= Debug; l++ {
			logMap[l] = openLog(*file + "." + logString[l])
		}
	} else {
		var _log *log.Logger
		if *file != "" {
			_log = openLog(*file)
		} else {
			_log = makeLogger(os.Stderr)
		}
		for l := Fatal; l <= Debug; l++ {
			logMap[l] = _log
		}
	}

	return New(logMap, lv, *only)
}

// You'll have to set up your own loggers for this one...
func New(m LogMap, lv LogLevel, only bool) *logger {
	// Sanity check the log map we've been passed.
	// We need loggers for all levels in case SetLogLevel is called.
	for l := Fatal; l <= Debug; l++ {
		if _log, ok := m[l]; !ok || _log == nil {
			log.Fatalf("Output log level %s has no logger configured.",
				logString[l])
		}
	}
	return &logger{m, lv, only, &sync.Mutex{}}
}

// Internal function all others call to ensure identical call depth
func (l *logger) write(lv LogLevel, fm string, v ...interface{}) {
	if lv > l.level || (l.only && lv != l.level) {
		// Your logs are not important to us, goodnight
		return
	}
	fm = fmt.Sprintf(LogString(lv)+" "+fm, v...)
	if _, ok := logString[lv]; !ok {
		// This is an unrecognised log level, so log it to Debug
		lv = Debug
	}

	l.Lock()
	defer l.Unlock()
	_log := l.log[lv]
	// Writing the log is deceptively simple
	_log.Output(3, fm)
	if lv == Fatal {
		// Always fatal to stderr too. Use panic so (a) we get a backtrace,
		// and (b) it's trappable for testing (and maybe other times too).
		log.Panic(fm)
	}
}

func (l *logger) Log(lv LogLevel, fm string, v ...interface{}) {
	l.write(lv, fm, v...)
}

// Helper functions for specific levels
func (l *logger) Debug(fm string, v ...interface{}) {
	l.write(Debug, fm, v...)
}

func (l *logger) Info(fm string, v ...interface{}) {
	l.write(Info, fm, v...)
}

func (l *logger) Warn(fm string, v ...interface{}) {
	l.write(Warn, fm, v...)
}

func (l *logger) Error(fm string, v ...interface{}) {
	l.write(Error, fm, v...)
}

func (l *logger) Fatal(fm string, v ...interface{}) {
	l.write(Fatal, fm, v...)
}

func (l *logger) SetLogLevel(lv LogLevel) {
	l.Lock()
	defer l.Unlock()
	l.level = lv
}

func (l *logger) SetOnly(only bool) {
	l.Lock()
	defer l.Unlock()
	l.only = only
}
