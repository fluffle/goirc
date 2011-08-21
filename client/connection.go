package client

import (
	"bufio"
	"crypto/tls"
	"github.com/fluffle/goirc/event"
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
// Once connected, any errors encountered are piped down *Conn.Err.
type Conn struct {
	// Connection Hostname and Nickname
	Host    string
	Me      *Nick
	Network string

	// Event handler registry and dispatcher
	Registry event.EventRegistry
	Dispatcher event.EventDispatcher

	// Map of channels we're on
	chans map[string]*Channel
	// Map of nicks we know about
	nicks map[string]*Nick

	// Use the State field to store external state that handlers might need.
	// Remember ... you might need locking for this ;-)
	State interface{}

	// I/O stuff to server
	sock      net.Conn
	io        *bufio.ReadWriter
	in        chan *Line
	out       chan string
	connected bool

	// Control channels to goroutines
	cSend, cLoop chan bool

	// Error channel to transmit any fail back to the user
	Err chan os.Error

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

	// Function which returns a *time.Time for use as a timestamp
	Timestamp func() *time.Time

	// Enable debugging? Set format for timestamps on debug output.
	Debug    bool
	TSFormat string
}

// Creates a new IRC connection object, but doesn't connect to anything so
// that you can add event handlers to it. See AddHandler() for details.
func New(nick, user, name string) *Conn {
	reg := event.NewRegistry()
	conn := &Conn{
		Registry: reg,
		Dispatcher: reg,
		in: make(chan *Line, 32),
		out: make(chan string, 32),
		Err: make(chan os.Error, 4),
		cSend: make(chan bool),
		cLoop: make(chan bool),
		SSL: false,
		SSLConfig: nil,
		Timeout: 300,
		Flood: false,
		badness: 0,
		lastsent: 0,
		Timestamp: time.LocalTime,
		TSFormat: "15:04:05",
	}
	conn.initialise()
	conn.SetupHandlers()
	conn.Me = conn.NewNick(nick, user, name, "")
	return conn
}

// Per-connection state initialisation.
func (conn *Conn) initialise() {
	conn.nicks = make(map[string]*Nick)
	conn.chans = make(map[string]*Channel)
	conn.io = nil
	conn.sock = nil

	// If this is being called because we are reconnecting, conn.Me
	// will still have all the old channels referenced -- nuke them!
	if conn.Me != nil {
		conn.Me = conn.NewNick(conn.Me.Nick, conn.Me.Ident, conn.Me.Name, "")
	}
}

// Connect the IRC connection object to "host[:port]" which should be either
// a hostname or an IP address, with an optional port. To enable explicit SSL
// on the connection to the IRC server, set Conn.SSL to true before calling
// Connect(). The port will default to 6697 if ssl is enabled, and 6667
// otherwise. You can also provide an optional connect password.
func (conn *Conn) Connect(host string, pass ...string) os.Error {
	if conn.connected {
		return os.NewError(fmt.Sprintf(
			"irc.Connect(): already connected to %s, cannot connect to %s",
			conn.Host, host))
	}

	if conn.SSL {
		if !hasPort(host) {
			host += ":6697"
		}
		if s, err := tls.Dial("tcp", host, conn.SSLConfig); err == nil {
			conn.sock = s
		} else {
			return err
		}
	} else {
		if !hasPort(host) {
			host += ":6667"
		}
		if s, err := net.Dial("tcp", host); err == nil {
			conn.sock = s
		} else {
			return err
		}
	}
	conn.Host = host
	conn.connected = true
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

// dispatch a nicely formatted os.Error to the error channel
func (conn *Conn) error(s string, a ...interface{}) {
	conn.Err <- os.NewError(fmt.Sprintf(s, a...))
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
			conn.error("irc.recv(): %s", err.String())
			conn.shutdown()
			break
		}
		s = strings.Trim(s, "\r\n")
		t := conn.Timestamp()
		if conn.Debug {
			fmt.Println(t.Format(conn.TSFormat) + " <- " + s)
		}

		if line := parseLine(s); line != nil {
			line.Time = t
			conn.in <- line
		}
	}
}

// goroutine to dispatch events for lines received on input channel
func (conn *Conn) runLoop() {
	for {
		select {
		case line := <-conn.in:
			conn.Dispatcher.Dispatch(line.Cmd, conn, line)
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
		conn.rateLimit(int64(len(line)))
	}

	if _, err := conn.io.WriteString(line + "\r\n"); err != nil {
		conn.error("irc.send(): %s", err.String())
		conn.shutdown()
		return
	}
	conn.io.Flush()
	if conn.Debug {
		fmt.Println(conn.Timestamp().Format(conn.TSFormat) + " -> " + line)
	}
}

// Implement Hybrid's flood control algorithm to rate-limit outgoing lines.
func (conn *Conn) rateLimit(chars int64) {
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
	if conn.badness > 10*second && !conn.Flood {
		// so sleep for the current line's time value before sending it
		time.Sleep(linetime)
	}
}

func (conn *Conn) shutdown() {
	// Guard against double-call of shutdown() if we get an error in send()
	// as calling sock.Close() will cause recv() to recieve EOF in readstring()
	if conn.connected {
		conn.connected = false
		conn.sock.Close()
		conn.cSend <- true
		conn.cLoop <- true
		conn.Dispatcher.Dispatch("disconnected", conn, &Line{})
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
	if conn.connected {
		str += "Connected to " + conn.Host + "\n\n"
	} else {
		str += "Not currently connected!\n\n"
	}
	str += conn.Me.String() + "\n"
	str += "GoIRC Channels\n"
	str += "--------------\n\n"
	for _, ch := range conn.chans {
		str += ch.String() + "\n"
	}
	str += "GoIRC NickNames\n"
	str += "---------------\n\n"
	for _, n := range conn.nicks {
		if n != conn.Me {
			str += n.String() + "\n"
		}
	}
	return str
}
