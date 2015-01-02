package glog

import (
	"fmt"
	"github.com/golang/glog"
	"github.com/fluffle/goirc/logging"
)

// Simple adapter to utilise Google's GLog package with goirc.
// Just import this package alongside goirc/client and call
// glog.Init() in your main() to set things up.
type GLogger struct{}

func (gl GLogger) Debug(f string, a ...interface{}) {
	// GLog doesn't have a "Debug" level, so use V(2) instead.
	if glog.V(2) {
		glog.InfoDepth(3, fmt.Sprintf(f, a...))
	}
}
func (gl GLogger) Info(f string, a ...interface{}) {
	glog.InfoDepth(3, fmt.Sprintf(f, a...))
}
func (gl GLogger) Warn(f string, a ...interface{}) {
	glog.WarningDepth(3, fmt.Sprintf(f, a...))
}
func (gl GLogger) Error(f string, a ...interface{}) {
	glog.ErrorDepth(3, fmt.Sprintf(f, a...))
}

func Init() {
	logging.SetLogger(GLogger{})
}
