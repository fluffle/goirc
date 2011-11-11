package client

import (
	"bufio"
	"github.com/fluffle/goirc/event"
	"github.com/fluffle/goirc/logging"
	"github.com/fluffle/goirc/state"
	"gomock.googlecode.com/hg/gomock"
	"testing"
	"time"
)

type testState struct {
	ctrl *gomock.Controller
	log  *logging.MockLogger
	st   *state.MockStateTracker
	ed   *event.MockEventDispatcher
	nc   *mockNetConn
	c    *Conn
}

func setUp(t *testing.T, start ...bool) (*Conn, *testState) {
	ctrl := gomock.NewController(t)
	st := state.NewMockStateTracker(ctrl)
	r := event.NewRegistry()
	ed := event.NewMockEventDispatcher(ctrl)
	l := logging.NewMockLogger(ctrl)
	nc := MockNetConn(t)
	c := Client("test", "test", "Testing IRC", r, l)

	// We don't want to have to specify s.log.EXPECT().Debug() for all the
	// random crap that gets logged. This mocks it all out nicely.
	ctrl.RecordCall(l, "Debug", gomock.Any(), gomock.Any()).AnyTimes()

	c.ED = ed
	c.ST = st
	c.st = true
	c.sock = nc
	c.Flood = true // Tests can take a while otherwise
	c.Connected = true
	if len(start) == 0 {
		// Hack to allow tests of send, recv, write etc.
		// NOTE: the value of the boolean doesn't matter.
		c.postConnect()
	}

	return c, &testState{ctrl, l, st, ed, nc, c}
}

func (s *testState) tearDown() {
	// This can get set to false in some tests
	s.c.st = true
	s.ed.EXPECT().Dispatch("disconnected", s.c, &Line{})
	s.st.EXPECT().Wipe()
	s.log.EXPECT().Error("irc.recv(): %s", "EOF")
	s.log.EXPECT().Info("irc.shutdown(): Disconnected from server.")
	s.nc.ExpectNothing()
	s.c.shutdown()
	<-time.After(1e6)
	s.ctrl.Finish()
}

// Practically the same as the above test, but shutdown is called implicitly
// by recv() getting an EOF from the mock connection.
func TestEOF(t *testing.T) {
	c, s := setUp(t)
	// Since we're not using tearDown() here, manually call Finish()
	defer s.ctrl.Finish()

	// Simulate EOF from server
	s.ed.EXPECT().Dispatch("disconnected", c, &Line{})
	s.st.EXPECT().Wipe()
	s.log.EXPECT().Info("irc.shutdown(): Disconnected from server.")
	s.log.EXPECT().Error("irc.recv(): %s", "EOF")
	s.nc.Close()

	// Since things happen in different internal goroutines, we need to wait
	// 1 ms should be enough :-)
	<-time.After(1e6)

	// Verify that the connection no longer thinks it's connected
	if c.Connected {
		t.Errorf("Conn still thinks it's connected to the server.")
	}
}

func TestClientAndStateTracking(t *testing.T) {
	// This doesn't use setUp() as we want to pass in a mock EventRegistry.
	ctrl := gomock.NewController(t)
	r := event.NewMockEventRegistry(ctrl)
	l := logging.NewMockLogger(ctrl)
	st := state.NewMockStateTracker(ctrl)

	for n, h := range intHandlers {
		r.EXPECT().AddHandler(h, n)
	}
	c := Client("test", "test", "Testing IRC", r, l)

	// Assert some basic things about the initial state of the Conn struct
	if c.ER != r || c.ED != r || c.l != l || c.st != false || c.ST != nil {
		t.Errorf("Conn not correctly initialised with external deps.")
	}
	if c.in == nil || c.out == nil || c.cSend == nil || c.cLoop == nil {
		t.Errorf("Conn control channels not correctly initialised.")
	}
	if c.Me.Nick != "test" || c.Me.Ident != "test" ||
		c.Me.Name != "Testing IRC" || c.Me.Host != "" {
		t.Errorf("Conn.Me not correctly initialised.")
	}

	// OK, while we're here with a mock event registry...
	for n, h := range stHandlers {
		r.EXPECT().AddHandler(h, n)
	}
	c.EnableStateTracking()

	// We're expecting the untracked me to be replaced by a tracked one.
	if c.Me.Nick != "test" || c.Me.Ident != "test" ||
		c.Me.Name != "Testing IRC" || c.Me.Host != "" {
		t.Errorf("Enabling state tracking did not replace Me correctly.")
	}
	if !c.st || c.ST == nil || c.Me != c.ST.Me() {
		t.Errorf("State tracker not enabled correctly.")
	}

	// Now, shim in the mock state tracker and test disabling state tracking.
	me := c.Me
	c.ST = st
	st.EXPECT().Wipe()
	for n, h := range stHandlers {
		r.EXPECT().DelHandler(h, n)
	}
	c.DisableStateTracking()
	if c.st || c.ST != nil || c.Me != me {
		t.Errorf("State tracker not disabled correctly.")
	}
	ctrl.Finish()
}

func TestSend(t *testing.T) {
	// Passing a second value to setUp inhibits postConnect()
	c, s := setUp(t, false)
	// We can't use tearDown here, as it will cause a deadlock in shutdown()
	// trying to send kill messages down channels to nonexistent goroutines.
	defer s.ctrl.Finish()

	// ... so we have to do some of it's work here.
	c.io = bufio.NewReadWriter(
		bufio.NewReader(c.sock),
		bufio.NewWriter(c.sock))

	// Assert that before send is running, nothing should be sent to the socket
	// but writes to the buffered channel "out" should not block.
	c.out <- "SENT BEFORE START"
	s.nc.ExpectNothing()

	// We want to test that the a goroutine calling send will exit correctly.
	exited := false
	go func() {
		c.send()
		exited = true
	}()

	// send is now running in the background as if started by postConnect.
	// This should read the line previously buffered in c.out, and write it
	// to the socket connection.
	s.nc.Expect("SENT BEFORE START")

	// Flood control is disabled -- setUp sets c.Flood = true -- so we should
	// not have set c.badness or c.lastsent. We test the actual rateLimit code
	// elsewhere, this is just for verification purposes...
	if c.badness != 0 || c.lastsent != 0 {
		t.Errorf("Flood control appears to be incorrectly enabled.")
	}

	// Send another line, just to be sure :-)
	c.out <- "SENT AFTER START"
	s.nc.Expect("SENT AFTER START")

	// Now, use the control channel to exit send and kill the goroutine.
	if exited {
		t.Errorf("Exited before signal sent.")
	}
	c.cSend <- true
	// Allow propagation time...
	<-time.After(1e6)
	if !exited {
		t.Errorf("Didn't exit after signal.")
	}
	s.nc.ExpectNothing()

	// Sending more on c.out shouldn't reach the network.
	c.out <- "SENT AFTER END"
	s.nc.ExpectNothing()
}
