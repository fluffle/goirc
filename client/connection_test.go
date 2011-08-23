package client

import (
	"strings"
	"testing"
	"time"
)

func setUp(t *testing.T) (*mockNetConn, *Conn) {
	c := New("test", "test", "Testing IRC")
	m := MockNetConn(t)
	c.sock = m
	c.Flood = true // Tests can take a while otherwise
	c.postConnect()
	return m, c
}

func tearDown(m *mockNetConn, c *Conn) {
	// This is enough to cause all the associated goroutines in m and c stop
	// (tested below in TestShutdown to make sure this is the case)
	m.Close()
}

func TestShutdown(t *testing.T) {
	_, c := setUp(t)

	// Shutdown won't clean things up unless we're "connected" already
	c.Connected = true

	// Setup a mock event dispatcher to test correct triggering of "disconnected"
	flag := false
	c.Dispatcher = WasEventDispatched("disconnected", &flag)

	// Call shutdown manually
	c.shutdown()

	// Check that we get an EOF from Read()
	timer := time.NewTimer(5e6)
	select {
	case <-timer.C:
		t.Errorf("No error received for shutdown.")
	case err := <-c.Err:
		timer.Stop()
		if err.String() != "irc.recv(): EOF" {
			t.Errorf("Expected EOF, got: %s", err)
		}
	}

	// Verify that the connection no longer thinks it's connected
	if c.Connected {
		t.Errorf("Conn still thinks it's connected to the server.")
	}

	// Verify that the "disconnected" event fired correctly
	if !flag {
		t.Errorf("Calling Close() didn't result in dispatch of disconnected event.")
	}

	// TODO(fluffle): Try to work out a way of testing that the background
	// goroutines were *actually* stopped? Test m a bit more?
}

// Practically the same as the above test, but shutdown is called implicitly
// by recv() getting an EOF from the mock connection.
func TestEOF(t *testing.T) {
	m, c := setUp(t)

	// Shutdown won't clean things up unless we're "connected" already
	c.Connected = true

	// Setup a mock event dispatcher to test correct triggering of "disconnected"
	flag := false
	c.Dispatcher = WasEventDispatched("disconnected", &flag)

	// Simulate EOF from server
	m.Close()

	// Check that we get an EOF from Read()
	timer := time.NewTimer(5e6)
	select {
	case <-timer.C:
		t.Errorf("No error received for shutdown.")
	case err := <-c.Err:
		timer.Stop()
		if err.String() != "irc.recv(): EOF" {
			t.Errorf("Expected EOF, got: %s", err)
		}
	}

	// Verify that the connection no longer thinks it's connected
	if c.Connected {
		t.Errorf("Conn still thinks it's connected to the server.")
	}

	// Verify that the "disconnected" event fired correctly
	if !flag {
		t.Errorf("Calling Close() didn't result in dispatch of disconnected event.")
	}
}

// Mock dispatcher to verify that events are triggered successfully
type mockDispatcher func(string, ...interface{})

func (d mockDispatcher) Dispatch(name string, ev ...interface{}) {
	d(name, ev...)
}

func WasEventDispatched(name string, flag *bool) mockDispatcher {
	return mockDispatcher(func(n string, ev ...interface{}) {
		if n == strings.ToLower(name) {
			*flag = true
		}
	})
}
