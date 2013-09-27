package logging

// The IRC client will log things using these methods
type Logger interface {
	// Debug logging of raw socket comms to/from server.
	Debug(format string, args ...interface{})
	// Informational logging about client behaviour.
	Info(format string, args ...interface{})
	// Warnings of inconsistent or unexpected data, mostly
	// related to state tracking of IRC nicks/chans.
	Warn(format string, args ...interface{})
	// Errors, mostly to do with network communication.
	Error(format string, args ...interface{})
}

// By default we do no logging. Logging is enabled or disabled
// at the package level, since I'm lazy and re-re-reorganising
// my code to pass a per-client-struct Logger around to all the
// state objects is a pain in the arse.
var logger Logger = nullLogger{}

// SetLogger sets the internal goirc Logger to l. If l is nil,
// a dummy logger that does nothing is installed instead.
func SetLogger(l Logger) {
	if l == nil {
		logger = nullLogger{}
	} else {
		logger = l
	}
}

// A nullLogger does nothing while fulfilling Logger.
type nullLogger struct{}
func (nl nullLogger) Debug(f string, a ...interface{}) {}
func (nl nullLogger) Info(f string, a ...interface{}) {}
func (nl nullLogger) Warn(f string, a ...interface{}) {}
func (nl nullLogger) Error(f string, a ...interface{}) {}

// Shim functions so that the package can be used directly
func Debug(f string, a ...interface{}) { logger.Debug(f, a...) }
func Info(f string, a ...interface{}) { logger.Info(f, a...) }
func Warn(f string, a ...interface{}) { logger.Warn(f, a...) }
func Error(f string, a ...interface{}) { logger.Error(f, a...) }
