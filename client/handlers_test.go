package client

import (
	"testing"
	"time"
)

// This test performs a simple end-to-end verification of correct line parsing
// and event dispatch as well as testing the PING handler. All the other tests
// in this file will call their respective handlers synchronously, otherwise
// testing becomes more difficult.
func TestPING(t *testing.T) {
	m, c := setUp(t)
	defer tearDown(m, c)
	m.Send("PING :1234567890")
	m.Expect("PONG :1234567890")
}

// Test the handler for 001 / RPL_WELCOME
func Test001(t *testing.T) {
	m, c := setUp(t)
	defer tearDown(m, c)

	// Setup a mock event dispatcher to test correct triggering of "connected"
	flag := false
	c.Dispatcher = WasEventDispatched("connected", &flag)

	// Assert that the "Host" field of c.Me hasn't been set yet
	if c.Me.Host != "" {
		t.Errorf("Host field contains unexpected value '%s'.", c.Me.Host)
	}
	
	// Call handler with a valid 001 line
	c.h_001(parseLine(":irc.server.org 001 test :Welcome to IRC test!ident@somehost.com"))
	// Should result in no response to server
	m.ExpectNothing()
	
	// Check that the event was dispatched correctly
	if !flag {
		t.Errorf("Sending 001 didn't result in dispatch of connected event.")
	}

	// Check host parsed correctly
	if c.Me.Host != "somehost.com" {
		t.Errorf("Host parsing failed, host is '%s'.", c.Me.Host)
	}
}

// Test the handler for 433 / ERR_NICKNAMEINUSE
func Test433(t *testing.T) {
	m, c := setUp(t)
	defer tearDown(m, c)
	
	// Assert that the nick set in setUp() is still "test" (just in case)
	if c.Me.Nick != "test" {
		t.Errorf("Tests will fail because Nick != 'test'.")
	}

	// Call handler with a 433 line, not triggering c.Me.Renick()
	c.h_433(parseLine(":irc.server.org 433 test new :Nickname is already in use."))
	m.Expect("NICK new_")

	// In this case, we're expecting the server to send a NICK line
	if c.Me.Nick != "test" {
		t.Errorf("ReNick() called unexpectedly, Nick == '%s'.", c.Me.Nick)
	}

	// Send a line that will trigger a renick. This happens when our wanted 
	// nick is unavailable during initial negotiation, so we must choose a
	// different one before the connection can proceed. No NICK line will be
	// sent by the server to confirm nick change in this case.
	c.h_433(parseLine(":irc.server.org 433 test test :Nickname is already in use."))
	m.Expect("NICK test_")

	if c.Me.Nick != "test_" {
		t.Errorf("ReNick() not called, Nick == '%s'.", c.Me.Nick)
	}
}

// Test the handler for NICK messages
func TestNICK(t *testing.T) {
	m, c := setUp(t)
	defer tearDown(m, c)

	// Assert that the nick set in setUp() is still "test" (just in case)
	if c.Me.Nick != "test" {
		t.Errorf("Tests will fail because Nick != 'test'.")
	}

	// Call handler with a NICK line changing "our" nick to test1.
	c.h_NICK(parseLine(":test!test@somehost.com NICK :test1"))
	// Should generate no response to server
	m.ExpectNothing()

	// Verify that our Nick has changed
	if c.Me.Nick != "test1" {
		t.Errorf("NICK did not result in changing our nick.")
	}

	// Create a "known" nick other than ours
	c.NewNick("user1", "ident1", "name one", "host1.com")

	// Call handler with a NICK line changing user1 to somebody
	c.h_NICK(parseLine(":user1!ident1@host1.com NICK :somebody"))
	if c.GetNick("user1") != nil {
		t.Errorf("Still have a valid Nick for 'user1'.")
	}
	if n := c.GetNick("somebody"); n == nil {
		t.Errorf("No Nick for 'somebody' found.")
	}

	// Send a NICK line for an unknown nick.
	c.h_NICK(parseLine(":blah!moo@cows.com NICK :milk"))

	// With the current error handling, we could block on reading the Err
	// channel, so ensure we don't wait forever with a 5ms timeout.
	timer := time.NewTimer(5e6)
	select {
	case <-timer.C:
		t.Errorf("No error received for bad NICK line.")
	case <-c.Err:
		timer.Stop()
	}
}

func TestCTCP(t *testing.T) {
	m, c := setUp(t)
	defer tearDown(m, c)

	// Call handler with CTCP VERSION
	c.h_CTCP(parseLine(":blah!moo@cows.com PRIVMSG test :\001VERSION\001"))

	// Expect a version reply
	m.Expect("NOTICE blah :\001VERSION powered by goirc...\001")

	// Call handler with CTCP PING
	c.h_CTCP(parseLine(":blah!moo@cows.com PRIVMSG test :\001PING 1234567890\001"))

	// Expect a ping reply
	m.Expect("NOTICE blah :\001PING 1234567890\001")

	// Call handler with CTCP UNKNOWN
	c.h_CTCP(parseLine(":blah!moo@cows.com PRIVMSG test :\001UNKNOWN ctcp\001"))

	// Expect nothing in reply
	m.ExpectNothing()
}
