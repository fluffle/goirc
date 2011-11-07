package client

import (
	"github.com/fluffle/goirc/event"
	"github.com/fluffle/goirc/logging"
	"github.com/fluffle/goirc/state"
	"gomock.googlecode.com/hg/gomock"
	"strings"
	"testing"
	"time"
)

type testState struct {
	ctrl *gomock.Controller
	log  *logging.MockLogger
	st   *state.MockStateTracker
	nc   *mockNetConn
	c    *Conn
}

func setUp(t *testing.T) (*Conn, *testState) {
	ctrl := gomock.NewController(t)
	st := state.NewMockStateTracker(ctrl)
	r := event.NewRegistry()
	l := logging.NewMockLogger(ctrl)
	nc := MockNetConn(t)
	c := Client("test", "test", "Testing IRC", r, l)

	// We don't want to have to specify s.log.EXPECT().Debug() for all the
	// random crap that gets logged. This mocks it all out nicely.
	ctrl.RecordCall(l, "Debug", gomock.Any(), gomock.Any()).AnyTimes()

	c.ST = st
	c.st = true
	c.sock = nc
	c.Flood = true // Tests can take a while otherwise
	c.Connected = true
	c.postConnect()

	// Assert some basic things about the initial state of the Conn struct
	if c.Me.Nick != "test" ||
		c.Me.Ident != "test" ||
		c.Me.Name != "Testing IRC" ||
		c.Me.Host != "" {
		t.Errorf("Conn.Me not correctly initialised.")
	}

	return c, &testState{ctrl, l, st, nc, c}
}

func (s *testState) tearDown() {
	// This can get set to false in some tests
	s.c.st = true
	s.st.EXPECT().Wipe()
	s.log.EXPECT().Error("irc.recv(): %s", "EOF")
	s.log.EXPECT().Info("irc.shutdown(): Disconnected from server.")
	s.nc.ExpectNothing()
	s.c.shutdown()
	<-time.After(1e6)
	s.ctrl.Finish()
}


func TestShutdown(t *testing.T) {
	c, s := setUp(t)

	// Setup a mock event dispatcher to test correct triggering of "disconnected"
	flag := c.ExpectEvent("disconnected")

	// Call shutdown via tearDown
	s.tearDown()

	// Verify that the connection no longer thinks it's connected
	if c.Connected {
		t.Errorf("Conn still thinks it's connected to the server.")
	}

	// Verify that the "disconnected" event fired correctly
	if !*flag {
		t.Errorf("Calling Close() didn't result in dispatch of disconnected event.")
	}

	// TODO(fluffle): Try to work out a way of testing that the background
	// goroutines were *actually* stopped? Test m a bit more?
}

// Practically the same as the above test, but shutdown is called implicitly
// by recv() getting an EOF from the mock connection.
func TestEOF(t *testing.T) {
	c, s := setUp(t)
	// Since we're not using tearDown() here, manually call Finish()
	defer s.ctrl.Finish()

	// Setup a mock event dispatcher to test correct triggering of "disconnected"
	flag := c.ExpectEvent("disconnected")

	// Simulate EOF from server
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

	// Verify that the "disconnected" event fired correctly
	if !*flag {
		t.Errorf("Calling Close() didn't result in dispatch of disconnected event.")
	}
}

// Mock dispatcher to verify that events are triggered successfully
type mockDispatcher func(string, ...interface{})

func (d mockDispatcher) Dispatch(name string, ev ...interface{}) {
	d(name, ev...)
}

func (conn *Conn) ExpectEvent(name string) *bool {
	flag := false
	conn.ED = mockDispatcher(func(n string, ev ...interface{}) {
		if n == strings.ToLower(name) {
			flag = true
		}
	})
	return &flag
}
