package client

import (
	"testing"
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
	// Should generate no errors and no response to server
	c.ExpectNoErrors()
	m.ExpectNothing()

	// Verify that our Nick has changed
	if c.Me.Nick != "test1" {
		t.Errorf("NICK did not result in changing our nick.")
	}

	// Create a "known" nick other than ours
	c.NewNick("user1", "ident1", "name one", "host1.com")

	// Call handler with a NICK line changing user1 to somebody
	c.h_NICK(parseLine(":user1!ident1@host1.com NICK :somebody"))
	c.ExpectNoErrors()
	m.ExpectNothing()

	if c.GetNick("user1") != nil {
		t.Errorf("Still have a valid Nick for 'user1'.")
	}
	if n := c.GetNick("somebody"); n == nil {
		t.Errorf("No Nick for 'somebody' found.")
	}

	// Send a NICK line for an unknown nick.
	c.h_NICK(parseLine(":blah!moo@cows.com NICK :milk"))
	c.ExpectError()
	m.ExpectNothing()
}

// Test the handler for CTCP messages
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

// Test the handler for JOIN messages
func TestJOIN(t *testing.T) {
	// TODO(fluffle): This tests a lot of extraneous functionality that should
	// be tested in nickchan_test. However, without mocking to ensure that
	// those functions are called correctly, we have to check they work by
	// verifying their expected side-effects instead. Fixing this requires
	// significant effort to move Conn to being a mockable interface type
	// instead of a concrete struct. I'm not sure how feasible this is :-/

	m, c := setUp(t)
	defer tearDown(m, c)

	// Assert that the nick set in setUp() is still "test" (just in case)
	if c.Me.Nick != "test" {
		t.Errorf("Tests will fail because Nick != 'test'.")
	}

	// Assert that we don't already know about our test channels
	if len(c.chans) != 0 {
		t.Errorf("Some channels are already known:")
		for _, ch := range c.chans {
			t.Logf(ch.String())
		}
	}

	// Use #test1 to test expected behaviour
	// Call handler with JOIN by test to #test1
	c.h_JOIN(parseLine(":test!test@somehost.com JOIN :#test1"))

	// Verify that we now know about #test1
	test1 := c.GetChannel("#test1")
	if test1 == nil || len(c.chans) != 1 {
		t.Errorf("Channel #test1 not tracked correctly:")
		for _, ch := range c.chans {
			t.Logf(ch.String())
		}
	}

	 // Verify we still only know about our own Nick
	if len(c.nicks) != 1 {
		t.Errorf("Other Nicks than ourselves exist:")
		for _, n := range c.nicks {
			t.Logf(n.String())
		}
	}

	// Verify that the channel has us and only in it
	if _, ok := test1.Nicks[c.Me]; !ok || len(test1.Nicks) != 1 {
		t.Errorf("Channel #test1 contains wrong nicks [1].")
	}

	// Verify that we're in the channel, and only that channel
	if _, ok := c.Me.Channels[test1]; !ok || len(c.Me.Channels) != 1 {
		t.Errorf("Nick (me) contains wrong channels.")
	}

	// Verify that the MODE and WHO commands are sent correctly
	m.Expect("MODE #test1")
	m.Expect("WHO #test1")

	// OK, now #test1 exists, JOIN another user we don't know about
	c.h_JOIN(parseLine(":user1!ident1@host1.com JOIN :#test1"))

	// Verify we created a new Nick for user1
	user1 := c.GetNick("user1")
	if user1 == nil || len(c.nicks) != 2 {
		t.Errorf("Unexpected number of known Nicks (wanted 2):")
		for _, n := range c.nicks {
			t.Logf(n.String())
		}
	}

	// Verify that test1 has us and user1 in, and that user1 is in test1.
	if _, ok := test1.Nicks[user1]; !ok || len(test1.Nicks) != 2 {
		t.Errorf("Channel #test1 contains wrong nicks [2].")
	}

	if _, ok := user1.Channels[test1]; !ok || len(user1.Channels) != 1 {
		t.Errorf("Nick user1 contains wrong channels.")
	}

	// Verify that the WHO command is sent correctly
	m.Expect("WHO user1")

	// Now, JOIN a nick we *do* know about.
	user2 := c.NewNick("user2", "ident2", "name two", "host2.com")

	// Ensure that user2 is in no channels beforehand, etc.
	if _, ok := test1.Nicks[user2]; ok || len(user2.Channels) != 0 {
		t.Errorf("Nick user2 in unexpected channels.")
	}

	c.h_JOIN(parseLine(":user2!ident2@host2.com JOIN :#test1"))
	if _, ok := test1.Nicks[user2]; !ok || len(test1.Nicks) != 3 {
		t.Errorf("Channel #test1 contains wrong nicks [3].")
	}
	if _, ok := user2.Channels[test1]; !ok || len(user2.Channels) != 1 {
		t.Errorf("Nick user2 contains wrong channels.")
	}
	// We already know about this user and channel, so nothing should be sent
	m.ExpectNothing()
}
