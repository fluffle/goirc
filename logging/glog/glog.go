package glog

import (
	"github.com/golang/glog"
)

// Simple adapter to utilise Google's GLog package with goirc.
// Just import github.com/fluffle/goirc/logging and this package,
// then call logging.Logger(glog.GLogger{}).
// The unfortunate downside of this is that it adds an extra hop
// to the caller chain which means *all* the line numbers are bad.
type GLogger struct{}

func (gl GLogger) Debug(f string, a ...interface{}) {
	// GLog doesn't have a "Debug" level, so use V(2) instead.
	glog.V(2).Infof(f, a...)
}
func (gl GLogger) Info(f string, a ...interface{}) { glog.Infof(f, a...) }
func (gl GLogger) Warn(f string, a ...interface{}) { glog.Warningf(f, a...) }
func (gl GLogger) Error(f string, a ...interface{}) { glog.Errorf(f, a...) }
