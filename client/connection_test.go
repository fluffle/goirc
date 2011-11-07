package client

import (
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

func setUp(t *testing.T) (*Conn, *testState) {
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
	c.postConnect()

	// Assert some basic things about the initial state of the Conn struct
	if c.Me.Nick != "test" ||
		c.Me.Ident != "test" ||
		c.Me.Name != "Testing IRC" ||
		c.Me.Host != "" {
		t.Errorf("Conn.Me not correctly initialised.")
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

func TestEnableStateTracking(t *testing.T) {

}
