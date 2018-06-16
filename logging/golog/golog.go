package golog

import (
	"github.com/fluffle/goirc/logging"
	log "github.com/fluffle/golog/logging"
)

// Simple adapter to utilise my logging package with goirc.
// Just import this package alongside goirc/client and call
// golog.Init() in your main() to set things up.
func Init() {
	l := log.NewFromFlags()
	l.SetDepth(1)
	logging.SetLogger(l)
}
