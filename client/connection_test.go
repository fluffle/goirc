package client

import (
	"github.com/fluffle/goirc/state"
	"github.com/golang/mock/gomock"
	"strings"
	"testing"
	"time"
)

type checker struct {
	t *testing.T
	c chan struct{}
}

func callCheck(t *testing.T) checker {
	return checker{t: t, c: make(chan struct{})}
}

func (c checker) call() {
	c.c <- struct{}{}
}

func (c checker) assertNotCalled(fmt string, args ...interface{}) {
	select {
	case <-c.c:
		c.t.Errorf(fmt, args...)
	default:
	}
}

func (c checker) assertWasCalled(fmt string, args ...interface{}) {
	select {
	case <-c.c:
	case <-time.After(time.Millisecond):
		// Usually need to wait for goroutines to settle :-/
		c.t.Errorf(fmt, args...)
	}
}

type testState struct {
	ctrl *gomock.Controller
	st   *state.MockTracker
	nc   *mockNetConn
	c    *Conn
}

// NOTE: including a second argument at all prevents calling c.postConnect()
func setUp(t *testing.T, start ...bool) (*Conn, *testState) {
	ctrl := gomock.NewController(t)
	st := state.NewMockTracker(ctrl)
	nc := MockNetConn(t)
	c := SimpleClient("test", "test", "Testing IRC")
	c.initialise()

	c.st = st
	c.sock = nc
	c.cfg.Flood = true // Tests can take a while otherwise
	c.connected = true
	// If a second argument is passed to setUp, we tell postConnect not to
	// start the various goroutines that shuttle data around.
	c.postConnect(len(start) == 0)
	// Sleep 1ms to allow background routines to start.
	<-time.After(time.Millisecond)

	return c, &testState{ctrl, st, nc, c}
}

func (s *testState) tearDown() {
	s.nc.ExpectNothing()
	s.c.shutdown()
	s.ctrl.Finish()
}

// Practically the same as the above test, but shutdown is called implicitly
// by recv() getting an EOF from the mock connection.
func TestEOF(t *testing.T) {
	c, s := setUp(t)
	// Since we're not using tearDown() here, manually call Finish()
	defer s.ctrl.Finish()

	// Set up a handler to detect whether disconnected handlers are called
	dcon := callCheck(t)
	c.HandleFunc(DISCONNECTED, func(conn *Conn, line *Line) {
		dcon.call()
	})

	// Simulate EOF from server
	s.nc.Close()

	// Verify that disconnected handler was called
	dcon.assertWasCalled("Conn did not call disconnected handlers.")

	// Verify that the connection no longer thinks it's connected
	if c.Connected() {
		t.Errorf("Conn still thinks it's connected to the server.")
	}
}

func TestClientAndStateTracking(t *testing.T) {
	ctrl := gomock.NewController(t)
	st := state.NewMockTracker(ctrl)
	c := SimpleClient("test", "test", "Testing IRC")

	// Assert some basic things about the initial state of the Conn struct
	me := c.cfg.Me
	if me.Nick != "test" || me.Ident != "test" ||
		me.Name != "Testing IRC" || me.Host != "" {
		t.Errorf("Conn.cfg.Me not correctly initialised.")
	}
	// Check that the internal handlers are correctly set up
	for k, _ := range intHandlers {
		if _, ok := c.intHandlers.set[strings.ToLower(k)]; !ok {
			t.Errorf("Missing internal handler for '%s'.", k)
		}
	}

	// Now enable the state tracking code and check its handlers
	c.EnableStateTracking()
	for k, _ := range stHandlers {
		if _, ok := c.intHandlers.set[strings.ToLower(k)]; !ok {
			t.Errorf("Missing state handler for '%s'.", k)
		}
	}
	if len(c.stRemovers) != len(stHandlers) {
		t.Errorf("Incorrect number of Removers (%d != %d) when adding state handlers.",
			len(c.stRemovers), len(stHandlers))
	}
	if neu := c.Me(); neu.Nick != me.Nick || neu.Ident != me.Ident ||
		neu.Name != me.Name || neu.Host != me.Host {
		t.Errorf("Enabling state tracking erased information about me!")
	}

	// We're expecting the untracked me to be replaced by a tracked one
	if c.st == nil {
		t.Errorf("State tracker not enabled correctly.")
	}
	if me = c.cfg.Me; me.Nick != "test" || me.Ident != "test" ||
		me.Name != "Testing IRC" || me.Host != "" {
		t.Errorf("Enabling state tracking did not replace Me correctly.")
	}

	// Now, shim in the mock state tracker and test disabling state tracking
	c.st = st
	gomock.InOrder(
		st.EXPECT().Me().Return(me),
		st.EXPECT().Wipe(),
	)
	c.DisableStateTracking()
	if c.st != nil || !c.cfg.Me.Equals(me) {
		t.Errorf("State tracker not disabled correctly.")
	}

	// Finally, check state tracking handlers were all removed correctly
	for k, _ := range stHandlers {
		if _, ok := c.intHandlers.set[strings.ToLower(k)]; ok && k != "NICK" {
			// A bit leaky, because intHandlers adds a NICK handler.
			t.Errorf("State handler for '%s' not removed correctly.", k)
		}
	}
	if len(c.stRemovers) != 0 {
		t.Errorf("stRemovers not zeroed correctly when removing state handlers.")
	}
	ctrl.Finish()
}

func TestSendExitsOnDie(t *testing.T) {
	// Passing a second value to setUp stops goroutines from starting
	c, s := setUp(t, false)
	defer s.tearDown()

	// Assert that before send is running, nothing should be sent to the socket
	// but writes to the buffered channel "out" should not block.
	c.out <- "SENT BEFORE START"
	s.nc.ExpectNothing()

	// We want to test that the a goroutine calling send will exit correctly.
	exited := callCheck(t)
	// send() will decrement the WaitGroup, so we must increment it.
	c.wg.Add(1)
	go func() {
		c.send()
		exited.call()
	}()

	// send is now running in the background as if started by postConnect.
	// This should read the line previously buffered in c.out, and write it
	// to the socket connection.
	s.nc.Expect("SENT BEFORE START")

	// Send another line, just to be sure :-)
	c.out <- "SENT AFTER START"
	s.nc.Expect("SENT AFTER START")

	// Now, use the control channel to exit send and kill the goroutine.
	// This sneakily uses the fact that the other two goroutines that would
	// normally be waiting for die to close are not running, so we only send
	// to the goroutine started above. Normally shutdown() closes c.die and
	// signals to all three goroutines (send, ping, runLoop) to exit.
	exited.assertNotCalled("Exited before signal sent.")
	c.die <- struct{}{}
	exited.assertWasCalled("Didn't exit after signal.")
	s.nc.ExpectNothing()

	// Sending more on c.out shouldn't reach the network.
	c.out <- "SENT AFTER END"
	s.nc.ExpectNothing()
}

func TestSendExitsOnWriteError(t *testing.T) {
	// Passing a second value to setUp stops goroutines from starting
	c, s := setUp(t, false)
	// We can't use tearDown here because we're testing shutdown conditions
	// (and so need to EXPECT() a call to st.Wipe() in the right place)
	defer s.ctrl.Finish()

	// We want to test that the a goroutine calling send will exit correctly.
	exited := callCheck(t)
	// send() will decrement the WaitGroup, so we must increment it.
	c.wg.Add(1)
	go func() {
		c.send()
		exited.call()
	}()

	// Send a line to be sure things are good.
	c.out <- "SENT AFTER START"
	s.nc.Expect("SENT AFTER START")

	// Now, close the underlying socket to cause write() to return an error.
	// This will call shutdown() => a call to st.Wipe() will happen.
	exited.assertNotCalled("Exited before signal sent.")
	s.nc.Close()
	// Sending more on c.out shouldn't reach the network, but we need to send
	// *something* to trigger a call to write() that will fail.
	c.out <- "SENT AFTER END"
	exited.assertWasCalled("Didn't exit after signal.")
	s.nc.ExpectNothing()
}

func TestRecv(t *testing.T) {
	// Passing a second value to setUp stops goroutines from starting
	c, s := setUp(t, false)
	// We can't use tearDown here because we're testing shutdown conditions
	// (and so need to EXPECT() a call to st.Wipe() in the right place)
	defer s.ctrl.Finish()

	// Send a line before recv is started up, to verify nothing appears on c.in
	s.nc.Send(":irc.server.org 001 test :First test line.")

	// reader is a helper to do a "non-blocking" read of c.in
	reader := func() *Line {
		select {
		case <-time.After(time.Millisecond):
		case l := <-c.in:
			return l
		}
		return nil
	}
	if l := reader(); l != nil {
		t.Errorf("Line parsed before recv started.")
	}

	// We want to test that the a goroutine calling recv will exit correctly.
	exited := callCheck(t)
	// recv() will decrement the WaitGroup, so we must increment it.
	c.wg.Add(1)
	go func() {
		c.recv()
		exited.call()
	}()

	// Now, this should mean that we'll receive our parsed line on c.in
	if l := reader(); l == nil || l.Cmd != "001" {
		t.Errorf("Bad first line received on input channel")
	}

	// Send a second line, just to be sure.
	s.nc.Send(":irc.server.org 002 test :Second test line.")
	if l := reader(); l == nil || l.Cmd != "002" {
		t.Errorf("Bad second line received on input channel.")
	}

	// Test that recv does something useful with a line it can't parse
	// (not that there are many, ParseLine is forgiving).
	s.nc.Send(":textwithnospaces")
	if l := reader(); l != nil {
		t.Errorf("Bad line still caused receive on input channel.")
	}

	// The only way recv() exits is when the socket closes.
	exited.assertNotCalled("Exited before socket close.")
	s.nc.Close()
	exited.assertWasCalled("Didn't exit on socket close.")

	// Since s.nc is closed we can't attempt another send on it...
	if l := reader(); l != nil {
		t.Errorf("Line received on input channel after socket close.")
	}
}

func TestPing(t *testing.T) {
	// Passing a second value to setUp stops goroutines from starting
	c, s := setUp(t, false)
	defer s.tearDown()

	// Set a low ping frequency for testing.
	c.cfg.PingFreq = 50 * time.Millisecond

	// reader is a helper to do a "non-blocking" read of c.out
	reader := func() string {
		select {
		case <-time.After(time.Millisecond):
		case s := <-c.out:
			return s
		}
		return ""
	}
	if s := reader(); s != "" {
		t.Errorf("Line output before ping started.")
	}

	// Start ping loop.
	exited := callCheck(t)
	// ping() will decrement the WaitGroup, so we must increment it.
	c.wg.Add(1)
	go func() {
		c.ping()
		exited.call()
	}()

	// The first ping should be after 50ms,
	// so we don't expect anything now on c.in
	if s := reader(); s != "" {
		t.Errorf("Line output directly after ping started.")
	}

	<-time.After(50 * time.Millisecond)
	if s := reader(); s == "" || !strings.HasPrefix(s, "PING :") {
		t.Errorf("Line not output after 50ms.")
	}

	// Reader waits for 1ms and we call it a few times above.
	<-time.After(45 * time.Millisecond)
	if s := reader(); s != "" {
		t.Errorf("Line output under 50ms after last ping.")
	}

	// This is a short window (49-51ms) in which the ping should happen
	// This may result in flaky tests; sorry (and file a bug) if so.
	<-time.After(2 * time.Millisecond)
	if s := reader(); s == "" || !strings.HasPrefix(s, "PING :") {
		t.Errorf("Line not output after another 2ms.")
	}

	// Now kill the ping loop.
	// This sneakily uses the fact that the other two goroutines that would
	// normally be waiting for die to close are not running, so we only send
	// to the goroutine started above. Normally shutdown() closes c.die and
	// signals to all three goroutines (send, ping, runLoop) to exit.
	exited.assertNotCalled("Exited before signal sent.")
	c.die <- struct{}{}
	exited.assertWasCalled("Didn't exit after signal.")
	// Make sure we're no longer pinging by waiting ~2x PingFreq
	<-time.After(105 * time.Millisecond)
	if s := reader(); s != "" {
		t.Errorf("Line output after ping stopped.")
	}
}

func TestRunLoop(t *testing.T) {
	// Passing a second value to setUp stops goroutines from starting
	c, s := setUp(t, false)
	defer s.tearDown()

	// Set up a handler to detect whether 001 handler is called
	h001 := callCheck(t)
	c.HandleFunc("001", func(conn *Conn, line *Line) {
		h001.call()
	})
	h002 := callCheck(t)
	// Set up a handler to detect whether 002 handler is called
	c.HandleFunc("002", func(conn *Conn, line *Line) {
		h002.call()
	})

	l1 := ParseLine(":irc.server.org 001 test :First test line.")
	c.in <- l1
	h001.assertNotCalled("001 handler called before runLoop started.")

	// We want to test that the a goroutine calling runLoop will exit correctly.
	// Now, we can expect the call to Dispatch to take place as runLoop starts.
	exited := callCheck(t)
	// runLoop() will decrement the WaitGroup, so we must increment it.
	c.wg.Add(1)
	go func() {
		c.runLoop()
		exited.call()
	}()
	h001.assertWasCalled("001 handler not called after runLoop started.")

	// Send another line, just to be sure :-)
	h002.assertNotCalled("002 handler called before expected.")
	l2 := ParseLine(":irc.server.org 002 test :Second test line.")
	c.in <- l2
	h002.assertWasCalled("002 handler not called while runLoop started.")

	// Now, use the control channel to exit send and kill the goroutine.
	// This sneakily uses the fact that the other two goroutines that would
	// normally be waiting for die to close are not running, so we only send
	// to the goroutine started above. Normally shutdown() closes c.die and
	// signals to all three goroutines (send, ping, runLoop) to exit.
	exited.assertNotCalled("Exited before signal sent.")
	c.die <- struct{}{}
	exited.assertWasCalled("Didn't exit after signal.")

	// Sending more on c.in shouldn't dispatch any further events
	c.in <- l1
	h001.assertNotCalled("001 handler called after runLoop ended.")
}

func TestWrite(t *testing.T) {
	// Passing a second value to setUp stops goroutines from starting
	c, s := setUp(t, false)
	// We can't use tearDown here because we're testing shutdown conditions
	// (and so need to EXPECT() a call to st.Wipe() in the right place)
	defer s.ctrl.Finish()

	// Write should just write a line to the socket.
	if err := c.write("yo momma"); err != nil {
		t.Errorf("Write returned unexpected error %v", err)
	}
	s.nc.Expect("yo momma")

	// Flood control is disabled -- setUp sets c.cfg.Flood = true -- so we should
	// not have set c.badness at this point.
	if c.badness != 0 {
		t.Errorf("Flood control used when Flood = true.")
	}

	c.cfg.Flood = false
	if err := c.write("she so useless"); err != nil {
		t.Errorf("Write returned unexpected error %v", err)
	}
	s.nc.Expect("she so useless")

	// The lastsent time should have been updated very recently...
	if time.Now().Sub(c.lastsent) > time.Millisecond {
		t.Errorf("Flood control not used when Flood = false.")
	}

	// Finally, test the error state by closing the socket then writing.
	s.nc.Close()
	if err := c.write("she can't pass unit tests"); err == nil {
		t.Errorf("Expected write to return error after socket close.")
	}
}

func TestRateLimit(t *testing.T) {
	c, s := setUp(t)
	defer s.tearDown()

	if c.badness != 0 {
		t.Errorf("Bad initial values for rate limit variables.")
	}

	// We'll be needing this later...
	abs := func(i time.Duration) time.Duration {
		if i < 0 {
			return -i
		}
		return i
	}

	// Since the changes to the time module, c.lastsent is now a time.Time.
	// It's initialised on client creation to time.Now() which for the purposes
	// of this test was probably around 1.2 ms ago. This is inconvenient.
	// Making it >10s ago effectively clears out the inconsistency, as this
	// makes elapsed > linetime and thus zeros c.badness and resets c.lastsent.
	c.lastsent = time.Now().Add(-10 * time.Second)
	if l := c.rateLimit(60); l != 0 || c.badness != 0 {
		t.Errorf("Rate limit got non-zero badness from long-ago lastsent.")
	}

	// So, time at the nanosecond resolution is a bit of a bitch. Choosing 60
	// characters as the line length means we should be increasing badness by
	// 2.5 seconds minus the delta between the two ratelimit calls. This should
	// be minimal but it's guaranteed that it won't be zero. Use 20us as a fuzz.
	if l := c.rateLimit(60); l != 0 ||
		abs(c.badness-2500*time.Millisecond) > 20*time.Microsecond {
		t.Errorf("Rate limit calculating badness incorrectly.")
	}
	// At this point, we can tip over the badness scale, with a bit of help.
	// 720 chars => +8 seconds of badness => 10.5 seconds => ratelimit
	if l := c.rateLimit(720); l != 8*time.Second ||
		abs(c.badness-10500*time.Millisecond) > 20*time.Microsecond {
		t.Errorf("Rate limit failed to return correct limiting values.")
		t.Errorf("l=%d, badness=%d", l, c.badness)
	}
}
