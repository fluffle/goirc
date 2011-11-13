package client

import (
	"bufio"
	"crypto/tls"
	"github.com/fluffle/goevent/event"
	"github.com/fluffle/golog/logging"
	"github.com/fluffle/goirc/state"
	"fmt"
	"net"
	"os"
	"strings"
	"time"
)

const (
	second = int64(1e9)
)

// An IRC connection is represented by this struct.
type Conn struct {
	// Connection Hostname and Nickname
	Host    string
	Me      *state.Nick
	Network string

	// Event handler registry and dispatcher
	ER event.EventRegistry
	ED event.EventDispatcher

	// State tracker for nicks and channels
	ST state.StateTracker
	st bool

	// Logger for debugging/warning/etc output
	l logging.Logger

	// Use the State field to store external state that handlers might need.
	// Remember ... you might need locking for this ;-)
	State interface{}

	// I/O stuff to server
	sock      net.Conn
	io        *bufio.ReadWriter
	in        chan *Line
	out       chan string
	Connected bool

	// Control channels to goroutines
	cSend, cLoop chan bool

	// Misc knobs to tweak client behaviour:
	// Are we connecting via SSL? Do we care about certificate validity?
	SSL       bool
	SSLConfig *tls.Config

	// Socket timeout, in seconds. Defaulted to 5m in New().
	Timeout int64

	// Set this to true to disable flood protection and false to re-enable
	Flood bool

	// Internal counters for flood protection
	badness, lastsent int64
}

// Creates a new IRC connection object, but doesn't connect to anything so
// that you can add event handlers to it. See AddHandler() for details.
func SimpleClient(nick string, args ...string) *Conn {
	r := event.NewRegistry()
	l := logging.NewFromFlags()
	ident := "goirc"
	name := "Powered by GoIRC"

	if len(args) > 0 && args[0] != "" {
		ident = args[0]
	}
	if len(args) > 1 && args[1] != "" {
		name = args[1]
	}
	return Client(nick, ident, name, r, l)
}

func Client(nick, ident, name string,
	r event.EventRegistry, l logging.Logger) *Conn {
	if r == nil || l == nil {
		return nil
	}
	conn := &Conn{
		ER:         r,
		ED:         r,
		l:          l,
		st:         false,
		in:         make(chan *Line, 32),
		out:        make(chan string, 32),
		cSend:      make(chan bool),
		cLoop:      make(chan bool),
		SSL:        false,
		SSLConfig:  nil,
		Timeout:    300,
		Flood:      false,
		badness:    0,
		lastsent:   0,
	}
	conn.addIntHandlers()
	conn.Me = state.NewNick(nick, l)
	conn.Me.Ident = ident
	conn.Me.Name = name

	conn.initialise()
	return conn
}

func (conn *Conn) EnableStateTracking() {
	if !conn.st {
		n := conn.Me
		conn.ST = state.NewTracker(n.Nick, conn.l)
		conn.Me = conn.ST.Me()
		conn.Me.Ident = n.Ident
		conn.Me.Name = n.Name
		conn.addSTHandlers()
		conn.st = true
	}
}

func (conn *Conn) DisableStateTracking() {
	if conn.st {
		conn.st = false
		conn.delSTHandlers()
		conn.ST.Wipe()
		conn.ST = nil
	}
}

// Per-connection state initialisation.
func (conn *Conn) initialise() {
	conn.io = nil
	conn.sock = nil
	if conn.st {
		conn.ST.Wipe()
	}
}

// Connect the IRC connection object to "host[:port]" which should be either
// a hostname or an IP address, with an optional port. To enable explicit SSL
// on the connection to the IRC server, set Conn.SSL to true before calling
// Connect(). The port will default to 6697 if ssl is enabled, and 6667
// otherwise. You can also provide an optional connect password.
func (conn *Conn) Connect(host string, pass ...string) os.Error {
	if conn.Connected {
		return os.NewError(fmt.Sprintf(
			"irc.Connect(): already connected to %s, cannot connect to %s",
			conn.Host, host))
	}

	if conn.SSL {
		if !hasPort(host) {
			host += ":6697"
		}
		conn.l.Info("irc.Connect(): Connecting to %s with SSL.", host)
		if s, err := tls.Dial("tcp", host, conn.SSLConfig); err == nil {
			conn.sock = s
		} else {
			return err
		}
	} else {
		if !hasPort(host) {
			host += ":6667"
		}
		conn.l.Info("irc.Connect(): Connecting to %s without SSL.", host)
		if s, err := net.Dial("tcp", host); err == nil {
			conn.sock = s
		} else {
			return err
		}
	}
	conn.Host = host
	conn.Connected = true
	conn.postConnect()

	if len(pass) > 0 {
		conn.Pass(pass[0])
	}
	conn.Nick(conn.Me.Nick)
	conn.User(conn.Me.Ident, conn.Me.Name)
	return nil
}

// Post-connection setup (for ease of testing)
func (conn *Conn) postConnect() {
	conn.io = bufio.NewReadWriter(
		bufio.NewReader(conn.sock),
		bufio.NewWriter(conn.sock))
	conn.sock.SetTimeout(conn.Timeout * second)
	go conn.send()
	go conn.recv()
	go conn.runLoop()
}

// copied from http.client for great justice
func hasPort(s string) bool {
	return strings.LastIndex(s, ":") > strings.LastIndex(s, "]")
}

// goroutine to pass data from output channel to write()
func (conn *Conn) send() {
	for {
		select {
		case line := <-conn.out:
			conn.write(line)
		case <-conn.cSend:
			// strobe on control channel, bail out
			return
		}
	}
}

// receive one \r\n terminated line from peer, parse and dispatch it
func (conn *Conn) recv() {
	for {
		s, err := conn.io.ReadString('\n')
		if err != nil {
			conn.l.Error("irc.recv(): %s", err.String())
			conn.shutdown()
			return
		}
		s = strings.Trim(s, "\r\n")
		conn.l.Debug("<- %s", s)

		if line := parseLine(s); line != nil {
			line.Time = time.LocalTime()
			conn.in <- line
		} else {
			conn.l.Warn("irc.recv(): problems parsing line:\n  %s", s)
		}
	}
}

// goroutine to dispatch events for lines received on input channel
func (conn *Conn) runLoop() {
	for {
		select {
		case line := <-conn.in:
			conn.ED.Dispatch(line.Cmd, conn, line)
		case <-conn.cLoop:
			// strobe on control channel, bail out
			return
		}
	}
}

// Write a \r\n terminated line of output to the connected server,
// using Hybrid's algorithm to rate limit if conn.Flood is false.
func (conn *Conn) write(line string) {
	if !conn.Flood {
		if t := conn.rateLimit(int64(len(line))); t != 0 {
			// sleep for the current line's time value before sending it
			conn.l.Debug("irc.rateLimit(): Flood! Sleeping for %.2f secs.",
				float64(t)/float64(second))
			<-time.After(t)
		}
	}

	if _, err := conn.io.WriteString(line + "\r\n"); err != nil {
		conn.l.Error("irc.send(): %s", err.String())
		conn.shutdown()
		return
	}
	if err := conn.io.Flush(); err != nil {
		conn.l.Error("irc.send(): %s", err.String())
		conn.shutdown()
		return
	}
	conn.l.Debug("-> %s", line)
}

// Implement Hybrid's flood control algorithm to rate-limit outgoing lines.
func (conn *Conn) rateLimit(chars int64) int64 {
	// Hybrid's algorithm allows for 2 seconds per line and an additional
	// 1/120 of a second per character on that line.
	linetime := 2*second + chars*second/120
	elapsed := time.Nanoseconds() - conn.lastsent
	if conn.badness += linetime - elapsed; conn.badness < 0 {
		// negative badness times are badness...
		conn.badness = int64(0)
	}
	conn.lastsent = time.Nanoseconds()
	// If we've sent more than 10 second's worth of lines according to the
	// calculation above, then we're at risk of "Excess Flood".
	if conn.badness > 10*second {
		return linetime
	}
	return 0
}

func (conn *Conn) shutdown() {
	// Guard against double-call of shutdown() if we get an error in send()
	// as calling sock.Close() will cause recv() to recieve EOF in readstring()
	if conn.Connected {
		conn.l.Info("irc.shutdown(): Disconnected from server.")
		conn.ED.Dispatch("disconnected", conn, &Line{})
		conn.Connected = false
		conn.sock.Close()
		conn.cSend <- true
		conn.cLoop <- true
		// reinit datastructures ready for next connection
		// do this here rather than after runLoop()'s for due to race
		conn.initialise()
	}
}

// Dumps a load of information about the current state of the connection to a
// string for debugging state tracking and other such things.
func (conn *Conn) String() string {
	str := "GoIRC Connection\n"
	str += "----------------\n\n"
	if conn.Connected {
		str += "Connected to " + conn.Host + "\n\n"
	} else {
		str += "Not currently connected!\n\n"
	}
	str += conn.Me.String() + "\n"
	if conn.st {
		str += conn.ST.String() + "\n"
	}
	return str
}
