package client

import (
	"github.com/fluffle/goirc/state"
	"github.com/golang/mock/gomock"
	"testing"
	"time"
)

// This test performs a simple end-to-end verification of correct line parsing
// and event dispatch as well as testing the PING handler. All the other tests
// in this file will call their respective handlers synchronously, otherwise
// testing becomes more difficult.
func TestPING(t *testing.T) {
	_, s := setUp(t)
	defer s.tearDown()
	s.nc.Send("PING :1234567890")
	s.nc.Expect("PONG :1234567890")
}

// Test the REGISTER handler matches section 3.1 of rfc2812
func TestREGISTER(t *testing.T) {
	c, s := setUp(t)
	defer s.tearDown()

	c.h_REGISTER(&Line{Cmd: REGISTER})
	s.nc.Expect("NICK test")
	s.nc.Expect("USER test 12 * :Testing IRC")
	s.nc.ExpectNothing()

	c.cfg.Pass = "12345"
	c.cfg.Me.Ident = "idiot"
	c.cfg.Me.Name = "I've got the same combination on my luggage!"
	c.h_REGISTER(&Line{Cmd: REGISTER})
	s.nc.Expect("PASS 12345")
	s.nc.Expect("NICK test")
	s.nc.Expect("USER idiot 12 * :I've got the same combination on my luggage!")
	s.nc.ExpectNothing()
}

// Test the handler for 001 / RPL_WELCOME
func Test001(t *testing.T) {
	c, s := setUp(t)
	defer s.tearDown()

	l := ParseLine(":irc.server.org 001 test :Welcome to IRC test!ident@somehost.com")
	// Set up a handler to detect whether connected handler is called from 001
	hcon := false
	c.HandleFunc("connected", func(conn *Conn, line *Line) {
		hcon = true
	})

	// Test state tracking first.
	gomock.InOrder(
		s.st.EXPECT().Me().Return(c.cfg.Me),
		s.st.EXPECT().NickInfo("test", "test", "somehost.com", "Testing IRC"),
	)
	// Call handler with a valid 001 line
	c.h_001(l)
	<-time.After(time.Millisecond)
	if !hcon {
		t.Errorf("001 handler did not dispatch connected event.")
	}

	// Now without state tracking.
	c.st = nil
	c.h_001(l)
	// Check host parsed correctly
	if c.cfg.Me.Host != "somehost.com" {
		t.Errorf("Host parsing failed, host is '%s'.", c.cfg.Me.Host)
	}
	c.st = s.st
}

// Test the handler for 433 / ERR_NICKNAMEINUSE
func Test433(t *testing.T) {
	c, s := setUp(t)
	defer s.tearDown()

	// Call handler with a 433 line, not triggering c.cfg.Me.Renick()
	s.st.EXPECT().Me().Return(c.cfg.Me)
	c.h_433(ParseLine(":irc.server.org 433 test new :Nickname is already in use."))
	s.nc.Expect("NICK new_")

	// Send a line that will trigger a renick. This happens when our wanted
	// nick is unavailable during initial negotiation, so we must choose a
	// different one before the connection can proceed. No NICK line will be
	// sent by the server to confirm nick change in this case.
	gomock.InOrder(
		s.st.EXPECT().Me().Return(c.cfg.Me),
		s.st.EXPECT().ReNick("test", "test_").Return(c.cfg.Me),
	)
	c.h_433(ParseLine(":irc.server.org 433 test test :Nickname is already in use."))
	s.nc.Expect("NICK test_")

	// Test the code path that *doesn't* involve state tracking.
	c.st = nil
	c.h_433(ParseLine(":irc.server.org 433 test test :Nickname is already in use."))
	s.nc.Expect("NICK test_")

	if c.cfg.Me.Nick != "test_" {
		t.Errorf("My nick not updated from '%s'.", c.cfg.Me.Nick)
	}
	c.st = s.st
}

// Test the handler for NICK messages when state tracking is disabled
func TestNICK(t *testing.T) {
	c, s := setUp(t)
	defer s.tearDown()

	// State tracking is enabled by default in setUp
	c.st = nil

	// Call handler with a NICK line changing "our" nick to test1.
	c.h_NICK(ParseLine(":test!test@somehost.com NICK :test1"))

	// Verify that our Nick has changed
	if c.cfg.Me.Nick != "test1" {
		t.Errorf("NICK did not result in changing our nick.")
	}

	// Send a NICK line for something that isn't us.
	c.h_NICK(ParseLine(":blah!moo@cows.com NICK :milk"))

	// Verify that our Nick hasn't changed
	if c.cfg.Me.Nick != "test1" {
		t.Errorf("NICK did not result in changing our nick.")
	}

	// Re-enable state tracking and send a line that *should* change nick.
	c.st = s.st
	c.h_NICK(ParseLine(":test1!test@somehost.com NICK :test2"))

	// Verify that our Nick hasn't changed (should be handled by h_STNICK).
	if c.cfg.Me.Nick != "test1" {
		t.Errorf("NICK changed our nick when state tracking enabled.")
	}
}

// Test the handler for CTCP messages
func TestCTCP(t *testing.T) {
	c, s := setUp(t)
	defer s.tearDown()

	// Call handler with CTCP VERSION
	c.h_CTCP(ParseLine(":blah!moo@cows.com PRIVMSG test :\001VERSION\001"))

	// Expect a version reply
	s.nc.Expect("NOTICE blah :\001VERSION Powered by GoIRC\001")

	// Call handler with CTCP PING
	c.h_CTCP(ParseLine(":blah!moo@cows.com PRIVMSG test :\001PING 1234567890\001"))

	// Expect a ping reply
	s.nc.Expect("NOTICE blah :\001PING 1234567890\001")

	// Call handler with CTCP UNKNOWN
	c.h_CTCP(ParseLine(":blah!moo@cows.com PRIVMSG test :\001UNKNOWN ctcp\001"))
}

// Test the handler for JOIN messages
func TestJOIN(t *testing.T) {
	c, s := setUp(t)
	defer s.tearDown()

	// The state tracker should be creating a new channel in this first test
	chan1 := &state.Channel{Name: "#test1"}

	gomock.InOrder(
		s.st.EXPECT().GetChannel("#test1").Return(nil),
		s.st.EXPECT().GetNick("test").Return(c.cfg.Me),
		s.st.EXPECT().Me().Return(c.cfg.Me),
		s.st.EXPECT().NewChannel("#test1").Return(chan1),
		s.st.EXPECT().Associate("#test1", "test"),
	)

	// Use #test1 to test expected behaviour
	// Call handler with JOIN by test to #test1
	c.h_JOIN(ParseLine(":test!test@somehost.com JOIN :#test1"))

	// Verify that the MODE and WHO commands are sent correctly
	s.nc.Expect("MODE #test1")
	s.nc.Expect("WHO #test1")

	// In this second test, we should be creating a new nick
	nick1 := &state.Nick{Nick: "user1"}

	gomock.InOrder(
		s.st.EXPECT().GetChannel("#test1").Return(chan1),
		s.st.EXPECT().GetNick("user1").Return(nil),
		s.st.EXPECT().NewNick("user1").Return(nick1),
		s.st.EXPECT().NickInfo("user1", "ident1", "host1.com", "").Return(nick1),
		s.st.EXPECT().Associate("#test1", "user1"),
	)

	// OK, now #test1 exists, JOIN another user we don't know about
	c.h_JOIN(ParseLine(":user1!ident1@host1.com JOIN :#test1"))

	// Verify that the WHO command is sent correctly
	s.nc.Expect("WHO user1")

	// In this third test, we'll be pretending we know about the nick already.
	nick2 := &state.Nick{Nick: "user2"}
	gomock.InOrder(
		s.st.EXPECT().GetChannel("#test1").Return(chan1),
		s.st.EXPECT().GetNick("user2").Return(nick2),
		s.st.EXPECT().Associate("#test1", "user2"),
	)
	c.h_JOIN(ParseLine(":user2!ident2@host2.com JOIN :#test1"))

	// Test error paths
	gomock.InOrder(
		// unknown channel, unknown nick
		s.st.EXPECT().GetChannel("#test2").Return(nil),
		s.st.EXPECT().GetNick("blah").Return(nil),
		s.st.EXPECT().Me().Return(c.cfg.Me),
		// unknown channel, known nick that isn't Me.
		s.st.EXPECT().GetChannel("#test2").Return(nil),
		s.st.EXPECT().GetNick("user2").Return(nick2),
		s.st.EXPECT().Me().Return(c.cfg.Me),
	)
	c.h_JOIN(ParseLine(":blah!moo@cows.com JOIN :#test2"))
	c.h_JOIN(ParseLine(":user2!ident2@host2.com JOIN :#test2"))
}

// Test the handler for PART messages
func TestPART(t *testing.T) {
	c, s := setUp(t)
	defer s.tearDown()

	// PART should dissociate a nick from a channel.
	s.st.EXPECT().Dissociate("#test1", "user1")
	c.h_PART(ParseLine(":user1!ident1@host1.com PART #test1 :Bye!"))
}

// Test the handler for KICK messages
// (this is very similar to the PART message test)
func TestKICK(t *testing.T) {
	c, s := setUp(t)
	defer s.tearDown()

	// KICK should dissociate a nick from a channel.
	s.st.EXPECT().Dissociate("#test1", "user1")
	c.h_KICK(ParseLine(":test!test@somehost.com KICK #test1 user1 :Bye!"))
}

// Test the handler for QUIT messages
func TestQUIT(t *testing.T) {
	c, s := setUp(t)
	defer s.tearDown()

	// Have user1 QUIT. All possible errors handled by state tracker \o/
	s.st.EXPECT().DelNick("user1")
	c.h_QUIT(ParseLine(":user1!ident1@host1.com QUIT :Bye!"))
}

// Test the handler for MODE messages
func TestMODE(t *testing.T) {
	c, s := setUp(t)
	defer s.tearDown()

	// Channel modes
	gomock.InOrder(
		s.st.EXPECT().GetChannel("#test1").Return(&state.Channel{Name: "#test1"}),
		s.st.EXPECT().ChannelModes("#test1", "+sk", "somekey"),
	)
	c.h_MODE(ParseLine(":user1!ident1@host1.com MODE #test1 +sk somekey"))

	// Nick modes for Me.
	gomock.InOrder(
		s.st.EXPECT().GetChannel("test").Return(nil),
		s.st.EXPECT().GetNick("test").Return(c.cfg.Me),
		s.st.EXPECT().Me().Return(c.cfg.Me),
		s.st.EXPECT().NickModes("test", "+i"),
	)
	c.h_MODE(ParseLine(":test!test@somehost.com MODE test +i"))

	// Check error paths
	gomock.InOrder(
		// send a valid user mode that's not us
		s.st.EXPECT().GetChannel("user1").Return(nil),
		s.st.EXPECT().GetNick("user1").Return(&state.Nick{Nick: "user1"}),
		s.st.EXPECT().Me().Return(c.cfg.Me),
		// Send a random mode for an unknown channel
		s.st.EXPECT().GetChannel("#test2").Return(nil),
		s.st.EXPECT().GetNick("#test2").Return(nil),
	)
	c.h_MODE(ParseLine(":user1!ident1@host1.com MODE user1 +w"))
	c.h_MODE(ParseLine(":user1!ident1@host1.com MODE #test2 +is"))
}

// Test the handler for TOPIC messages
func TestTOPIC(t *testing.T) {
	c, s := setUp(t)
	defer s.tearDown()

	// Ensure TOPIC reply calls Topic
	gomock.InOrder(
		s.st.EXPECT().GetChannel("#test1").Return(&state.Channel{Name: "#test1"}),
		s.st.EXPECT().Topic("#test1", "something something"),
	)
	c.h_TOPIC(ParseLine(":user1!ident1@host1.com TOPIC #test1 :something something"))

	// Check error paths -- send a topic for an unknown channel
	s.st.EXPECT().GetChannel("#test2").Return(nil)
	c.h_TOPIC(ParseLine(":user1!ident1@host1.com TOPIC #test2 :dark side"))
}

// Test the handler for 311 / RPL_WHOISUSER
func Test311(t *testing.T) {
	c, s := setUp(t)
	defer s.tearDown()

	// Ensure 311 reply calls NickInfo
	gomock.InOrder(
		s.st.EXPECT().GetNick("user1").Return(&state.Nick{Nick: "user1"}),
		s.st.EXPECT().Me().Return(c.cfg.Me),
		s.st.EXPECT().NickInfo("user1", "ident1", "host1.com", "name"),
	)
	c.h_311(ParseLine(":irc.server.org 311 test user1 ident1 host1.com * :name"))

	// Check error paths -- send a 311 for an unknown nick
	s.st.EXPECT().GetNick("user2").Return(nil)
	c.h_311(ParseLine(":irc.server.org 311 test user2 ident2 host2.com * :dongs"))
}

// Test the handler for 324 / RPL_CHANNELMODEIS
func Test324(t *testing.T) {
	c, s := setUp(t)
	defer s.tearDown()

	// Ensure 324 reply calls ChannelModes
	gomock.InOrder(
		s.st.EXPECT().GetChannel("#test1").Return(&state.Channel{Name: "#test1"}),
		s.st.EXPECT().ChannelModes("#test1", "+sk", "somekey"),
	)
	c.h_324(ParseLine(":irc.server.org 324 test #test1 +sk somekey"))

	// Check error paths -- send 324 for an unknown channel
	s.st.EXPECT().GetChannel("#test2").Return(nil)
	c.h_324(ParseLine(":irc.server.org 324 test #test2 +pmt"))
}

// Test the handler for 332 / RPL_TOPIC
func Test332(t *testing.T) {
	c, s := setUp(t)
	defer s.tearDown()

	// Ensure 332 reply calls Topic
	gomock.InOrder(
		s.st.EXPECT().GetChannel("#test1").Return(&state.Channel{Name: "#test1"}),
		s.st.EXPECT().Topic("#test1", "something something"),
	)
	c.h_332(ParseLine(":irc.server.org 332 test #test1 :something something"))

	// Check error paths -- send 332 for an unknown channel
	s.st.EXPECT().GetChannel("#test2").Return(nil)
	c.h_332(ParseLine(":irc.server.org 332 test #test2 :dark side"))
}

// Test the handler for 352 / RPL_WHOREPLY
func Test352(t *testing.T) {
	c, s := setUp(t)
	defer s.tearDown()

	// Ensure 352 reply calls NickInfo and NickModes
	gomock.InOrder(
		s.st.EXPECT().GetNick("user1").Return(&state.Nick{Nick: "user1"}),
		s.st.EXPECT().Me().Return(c.cfg.Me),
		s.st.EXPECT().NickInfo("user1", "ident1", "host1.com", "name"),
	)
	c.h_352(ParseLine(":irc.server.org 352 test #test1 ident1 host1.com irc.server.org user1 G :0 name"))

	// Check that modes are set correctly from WHOREPLY
	gomock.InOrder(
		s.st.EXPECT().GetNick("user1").Return(&state.Nick{Nick: "user1"}),
		s.st.EXPECT().Me().Return(c.cfg.Me),
		s.st.EXPECT().NickInfo("user1", "ident1", "host1.com", "name"),
		s.st.EXPECT().NickModes("user1", "+o"),
		s.st.EXPECT().NickModes("user1", "+i"),
	)
	c.h_352(ParseLine(":irc.server.org 352 test #test1 ident1 host1.com irc.server.org user1 H* :0 name"))

	// Check error paths -- send a 352 for an unknown nick
	s.st.EXPECT().GetNick("user2").Return(nil)
	c.h_352(ParseLine(":irc.server.org 352 test #test2 ident2 host2.com irc.server.org user2 G :0 fooo"))
}

// Test the handler for 353 / RPL_NAMREPLY
func Test353(t *testing.T) {
	c, s := setUp(t)
	defer s.tearDown()

	// 353 handler is called twice, so GetChannel will be called twice
	s.st.EXPECT().GetChannel("#test1").Return(&state.Channel{Name: "#test1"}).Times(2)
	gomock.InOrder(
		// "test" is Me, i am known, and already on the channel
		s.st.EXPECT().GetNick("test").Return(c.cfg.Me),
		s.st.EXPECT().IsOn("#test1", "test").Return(&state.ChanPrivs{}, true),
		// user1 is known, but not on the channel, so should be associated
		s.st.EXPECT().GetNick("user1").Return(&state.Nick{Nick: "user1"}),
		s.st.EXPECT().IsOn("#test1", "user1").Return(nil, false),
		s.st.EXPECT().Associate("#test1", "user1").Return(&state.ChanPrivs{}),
		s.st.EXPECT().ChannelModes("#test1", "+o", "user1"),
	)
	for n, m := range map[string]string{
		"user2":  "",
		"voice":  "+v",
		"halfop": "+h",
		"op":     "+o",
		"admin":  "+a",
		"owner":  "+q",
	} {
		calls := []*gomock.Call{
			s.st.EXPECT().GetNick(n).Return(nil),
			s.st.EXPECT().NewNick(n).Return(&state.Nick{Nick: n}),
			s.st.EXPECT().IsOn("#test1", n).Return(nil, false),
			s.st.EXPECT().Associate("#test1", n).Return(&state.ChanPrivs{}),
		}
		if m != "" {
			calls = append(calls, s.st.EXPECT().ChannelModes("#test1", m, n))
		}
		gomock.InOrder(calls...)
	}

	// Send a couple of names replies (complete with trailing space)
	c.h_353(ParseLine(":irc.server.org 353 test = #test1 :test @user1 user2 +voice "))
	c.h_353(ParseLine(":irc.server.org 353 test = #test1 :%halfop @op &admin ~owner "))

	// Check error paths -- send 353 for an unknown channel
	s.st.EXPECT().GetChannel("#test2").Return(nil)
	c.h_353(ParseLine(":irc.server.org 353 test = #test2 :test ~user3"))
}

// Test the handler for 671 (unreal specific)
func Test671(t *testing.T) {
	c, s := setUp(t)
	defer s.tearDown()

	// Ensure 671 reply calls NickModes
	gomock.InOrder(
		s.st.EXPECT().GetNick("user1").Return(&state.Nick{Nick: "user1"}),
		s.st.EXPECT().NickModes("user1", "+z"),
	)
	c.h_671(ParseLine(":irc.server.org 671 test user1 :some ignored text"))

	// Check error paths -- send a 671 for an unknown nick
	s.st.EXPECT().GetNick("user2").Return(nil)
	c.h_671(ParseLine(":irc.server.org 671 test user2 :some ignored text"))
}
