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
	user1 := c.NewNick("user1", "ident1", "name one", "host1.com")

	// Call handler with a NICK line changing user1 to somebody
	c.h_NICK(parseLine(":user1!ident1@host1.com NICK :somebody"))
	c.ExpectNoErrors()
	m.ExpectNothing()

	if c.GetNick("user1") != nil {
		t.Errorf("Still have a valid Nick for 'user1'.")
	}
	if n := c.GetNick("somebody"); n != user1 {
		t.Errorf("GetNick(somebody) didn't result in correct Nick.")
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
	// TODO(fluffle): Without mocking to ensure that the various methods
	// h_JOIN uses are called, we must check they do the right thing by
	// verifying their expected side-effects instead. Fixing this requires
	// significant effort to move Conn to being a mockable interface type
	// instead of a concrete struct. I'm not sure how feasible this is :-/
	//
	// Instead, in this test we (so far) just verify the correct code paths
	// are followed and trust that the unit tests for the various methods
	// ensure that they do the right thing.

	m, c := setUp(t)
	defer tearDown(m, c)

	// Use #test1 to test expected behaviour
	// Call handler with JOIN by test to #test1
	c.h_JOIN(parseLine(":test!test@somehost.com JOIN :#test1"))

	// Ensure we aren't triggering an error here
	c.ExpectNoErrors()

	// Verify that the MODE and WHO commands are sent correctly
	m.Expect("MODE #test1")
	m.Expect("WHO #test1")

	// Simple verification that NewChannel was called for #test1
	test1 := c.GetChannel("#test1")
	if test1 == nil {
		t.Errorf("No Channel for #test1 created on JOIN.")
	}

	// OK, now #test1 exists, JOIN another user we don't know about
	c.h_JOIN(parseLine(":user1!ident1@host1.com JOIN :#test1"))

	// Again, expect no errors
	c.ExpectNoErrors()

	// Verify that the WHO command is sent correctly
	m.Expect("WHO user1")

	// Simple verification that NewNick was called for user1
	user1 := c.GetNick("user1")
	if user1 == nil {
		t.Errorf("No Nick for user1 created on JOIN.")
	}

	// Now, JOIN a nick we *do* know about.
	user2 := c.NewNick("user2", "ident2", "name two", "host2.com")
	c.h_JOIN(parseLine(":user2!ident2@host2.com JOIN :#test1"))

	// We already know about this user and channel, so nothing should be sent
	c.ExpectNoErrors()
	m.ExpectNothing()

	// Simple verification that the state tracking has actually been done
	if _, ok := test1.Nicks[user2]; !ok || len(test1.Nicks) != 3 {
		t.Errorf("State tracking horked, hopefully other unit tests fail.")
	}

	// Test error paths -- unknown channel, unknown nick
	c.h_JOIN(parseLine(":blah!moo@cows.com JOIN :#test2"))
	c.ExpectError()
	m.ExpectNothing()

	// unknown channel, known nick that isn't Me.
	c.h_JOIN(parseLine(":user2!ident2@host2.com JOIN :#test2"))
	c.ExpectError()
	m.ExpectNothing()
}

// Test the handler for PART messages
func TestPART(t *testing.T) {
	m, c := setUp(t)
	defer tearDown(m, c)

	// Create user1 and add them to #test1 and #test2
	user1 := c.NewNick("user1", "ident1", "name one", "host1.com")
	test1 := c.NewChannel("#test1")
	test2 := c.NewChannel("#test2")
	test1.AddNick(user1)
	test2.AddNick(user1)

	// Add Me to both channels (not strictly necessary)
	test1.AddNick(c.Me)
	test2.AddNick(c.Me)

	// Then make them PART
	c.h_PART(parseLine(":user1!ident1@host1.com PART #test1 :Bye!"))

	// Expect no errors or output
	c.ExpectNoErrors()
	m.ExpectNothing()

	// Quick check of tracking code
	if len(test1.Nicks) != 1 {
		t.Errorf("PART failed to remove user1 from #test1.")
	}

	// Test error states.
	// Part a known user from a known channel they are not on.
	c.h_PART(parseLine(":user1!ident1@host1.com PART #test1 :Bye!"))
	c.ExpectError()

	// Part an unknown user from a known channel.
	c.h_PART(parseLine(":user2!ident2@host2.com PART #test1 :Bye!"))
	c.ExpectError()

	// Part a known user from an unknown channel.
	c.h_PART(parseLine(":user1!ident1@host1.com PART #test3 :Bye!"))
	c.ExpectError()

	// Part an unknown user from an unknown channel.
	c.h_PART(parseLine(":user2!ident2@host2.com PART #test3 :Bye!"))
	c.ExpectError()
}

// Test the handler for KICK messages
// (this is very similar to the PART message test)
func TestKICK(t *testing.T) {
	m, c := setUp(t)
	defer tearDown(m, c)

	// Create user1 and add them to #test1 and #test2
	user1 := c.NewNick("user1", "ident1", "name one", "host1.com")
	test1 := c.NewChannel("#test1")
	test2 := c.NewChannel("#test2")
	test1.AddNick(user1)
	test2.AddNick(user1)

	// Add Me to both channels (not strictly necessary)
	test1.AddNick(c.Me)
	test2.AddNick(c.Me)

	// Then kick them!
	c.h_KICK(parseLine(":test!test@somehost.com KICK #test1 user1 :Bye!"))

	// Expect no errors or output
	c.ExpectNoErrors()
	m.ExpectNothing()

	// Quick check of tracking code
	if len(test1.Nicks) != 1 {
		t.Errorf("PART failed to remove user1 from #test1.")
	}

	// Test error states.
	// Kick a known user from a known channel they are not on.
	c.h_KICK(parseLine(":test!test@somehost.com KICK #test1 user1 :Bye!"))
	c.ExpectError()

	// Kick an unknown user from a known channel.
	c.h_KICK(parseLine(":test!test@somehost.com KICK #test2 user2 :Bye!"))
	c.ExpectError()

	// Kick a known user from an unknown channel.
	c.h_KICK(parseLine(":test!test@somehost.com KICK #test3 user1 :Bye!"))
	c.ExpectError()

	// Kick an unknown user from an unknown channel.
	c.h_KICK(parseLine(":test!test@somehost.com KICK #test4 user2 :Bye!"))
	c.ExpectError()
}
