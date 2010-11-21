package client

import (
	"testing"
)

// Not really sure what or how to test something that basically requires a
// connection to an IRC server to function, but we need some tests or when this
// is present in the go package tree builds fail hard :-(
func TestIRC(t *testing.T) {
	if c := New("test", "test", "Testing IRC"); c == nil {
		t.FailNow()
	}
}
