package client

import (
	"bufio"
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/fluffle/goirc/state"
	"github.com/fluffle/golog/logging"
	"net"
	"strings"
	"time"
)

// An IRC connection is represented by this struct.
type Conn struct {
	// Connection related vars people will care about
	Me        *state.Nick
	Host      string
	Network   string
	Connected bool

	// Deprecated: future work to turn Conn into an interface will break this.
	// Use the State field to store external state that handlers might need.
	State interface{}

	// Contains parameters that people can tweak to change client behaviour.
	cfg *Config

	// Handlers and Commands
	handlers *handlerSet
	commands *commandList

	// State tracker for nicks and channels
	st         state.Tracker
	stRemovers []Remover

	// I/O stuff to server
	sock net.Conn
	io   *bufio.ReadWriter
	in   chan *Line
	out  chan string

	// Control channels to goroutines
	cSend, cLoop, cPing chan bool

	// Internal counters for flood protection
	badness  time.Duration
	lastsent time.Time
}

// Misc knobs to tweak client behaviour go in here
type Config struct {
	// Set this to provide the Nick, Ident and Name for the client to use.
	Me *state.Nick

	// Are we connecting via SSL? Do we care about certificate validity?
	SSL       bool
	SSLConfig *tls.Config

	// Replaceable function to customise the 433 handler's new nick
	NewNick func(string) string

	// Client->server ping frequency, in seconds. Defaults to 3m.
	PingFreq time.Duration

	// Controls what is stripped from line.Args[1] for Commands
	CommandStripNick, SimpleCommandStripPrefix bool

	// Set this to true to disable flood protection and false to re-enable
	Flood bool
}

func NewConfig(nick string, args ...string) *Config {
	ident := "goirc"
	name := "Powered by GoIRC"

	if len(args) > 0 && args[0] != "" {
		ident = args[0]
	}
	if len(args) > 1 && args[1] != "" {
		name = args[1]
	}
	cfg := &Config{
		PingFreq: 3 * time.Minute,
		NewNick:  func(s string) string { return s + "_" },
	}
	cfg.Me = state.NewNick(nick)
	cfg.Me.Ident = ident
	cfg.Me.Name = name
	return cfg
}

// Creates a new IRC connection object, but doesn't connect to anything so
// that you can add event handlers to it. See AddHandler() for details
func SimpleClient(nick string, args ...string) (*Conn, error) {
	return Client(NewConfig(nick, args...))
}

func Client(cfg *Config) (*Conn, error) {
	logging.InitFromFlags()
	if cfg.Me == nil || cfg.Me.Nick == "" || cfg.Me.Ident == "" {
		return nil, fmt.Errorf("Must provide a valid state.Nick in cfg.Me.")
	}
	conn := &Conn{
		Me:         cfg.Me,
		cfg:        cfg,
		in:         make(chan *Line, 32),
		out:        make(chan string, 32),
		cSend:      make(chan bool),
		cLoop:      make(chan bool),
		cPing:      make(chan bool),
		handlers:   newHandlerSet(),
		commands:   newCommandList(),
		stRemovers: make([]Remover, 0, len(stHandlers)),
		lastsent:   time.Now(),
	}
	conn.addIntHandlers()
	conn.initialise()
	return conn, nil
}

func (conn *Conn) Config() *Config {
	return conn.cfg
}

func (conn *Conn) EnableStateTracking() {
	if conn.st == nil {
		n := conn.Me
		conn.st = state.NewTracker(n.Nick)
		conn.Me = conn.st.Me()
		conn.Me.Ident = n.Ident
		conn.Me.Name = n.Name
		conn.addSTHandlers()
	}
}

func (conn *Conn) DisableStateTracking() {
	if conn.st != nil {
		conn.delSTHandlers()
		conn.st.Wipe()
		conn.st = nil
	}
}

func (conn *Conn) StateTracker() state.Tracker {
	return conn.st
}

// Per-connection state initialisation.
func (conn *Conn) initialise() {
	conn.io = nil
	conn.sock = nil
	if conn.st != nil {
		conn.st.Wipe()
	}
}

// Connect the IRC connection object to "host[:port]" which should be either
// a hostname or an IP address, with an optional port. To enable explicit SSL
// on the connection to the IRC server, set Conn.SSL to true before calling
// Connect(). The port will default to 6697 if ssl is enabled, and 6667
// otherwise. You can also provide an optional connect password.
func (conn *Conn) Connect(host string, pass ...string) error {
	if conn.Connected {
		return errors.New(fmt.Sprintf(
			"irc.Connect(): already connected to %s, cannot connect to %s",
			conn.Host, host))
	}

	if conn.cfg.SSL {
		if !hasPort(host) {
			host += ":6697"
		}
		logging.Info("irc.Connect(): Connecting to %s with SSL.", host)
		if s, err := tls.Dial("tcp", host, conn.cfg.SSLConfig); err == nil {
			conn.sock = s
		} else {
			return err
		}
	} else {
		if !hasPort(host) {
			host += ":6667"
		}
		logging.Info("irc.Connect(): Connecting to %s without SSL.", host)
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
	go conn.send()
	go conn.recv()
	if conn.cfg.PingFreq > 0 {
		go conn.ping()
	} else {
		// Otherwise the send in shutdown will hang :-/
		go func() { <-conn.cPing }()
	}
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
			logging.Error("irc.recv(): %s", err.Error())
			conn.shutdown()
			return
		}
		s = strings.Trim(s, "\r\n")
		logging.Debug("<- %s", s)

		if line := parseLine(s); line != nil {
			line.Time = time.Now()
			conn.in <- line
		} else {
			logging.Warn("irc.recv(): problems parsing line:\n  %s", s)
		}
	}
}

// Repeatedly pings the server every PingFreq seconds (no matter what)
func (conn *Conn) ping() {
	tick := time.NewTicker(conn.cfg.PingFreq)
	for {
		select {
		case <-tick.C:
			conn.Raw(fmt.Sprintf("PING :%d", time.Now().UnixNano()))
		case <-conn.cPing:
			tick.Stop()
			return
		}
	}
}

// goroutine to dispatch events for lines received on input channel
func (conn *Conn) runLoop() {
	for {
		select {
		case line := <-conn.in:
			conn.dispatch(line)
		case <-conn.cLoop:
			// strobe on control channel, bail out
			return
		}
	}
}

// Write a \r\n terminated line of output to the connected server,
// using Hybrid's algorithm to rate limit if conn.cfg.Flood is false.
func (conn *Conn) write(line string) {
	if !conn.cfg.Flood {
		if t := conn.rateLimit(len(line)); t != 0 {
			// sleep for the current line's time value before sending it
			logging.Debug("irc.rateLimit(): Flood! Sleeping for %.2f secs.",
				t.Seconds())
			<-time.After(t)
		}
	}

	if _, err := conn.io.WriteString(line + "\r\n"); err != nil {
		logging.Error("irc.send(): %s", err.Error())
		conn.shutdown()
		return
	}
	if err := conn.io.Flush(); err != nil {
		logging.Error("irc.send(): %s", err.Error())
		conn.shutdown()
		return
	}
	logging.Debug("-> %s", line)
}

// Implement Hybrid's flood control algorithm to rate-limit outgoing lines.
func (conn *Conn) rateLimit(chars int) time.Duration {
	// Hybrid's algorithm allows for 2 seconds per line and an additional
	// 1/120 of a second per character on that line.
	linetime := 2*time.Second + time.Duration(chars)*time.Second/120
	elapsed := time.Now().Sub(conn.lastsent)
	if conn.badness += linetime - elapsed; conn.badness < 0 {
		// negative badness times are badness...
		conn.badness = 0
	}
	conn.lastsent = time.Now()
	// If we've sent more than 10 second's worth of lines according to the
	// calculation above, then we're at risk of "Excess Flood".
	if conn.badness > 10*time.Second {
		return linetime
	}
	return 0
}

func (conn *Conn) shutdown() {
	// Guard against double-call of shutdown() if we get an error in send()
	// as calling sock.Close() will cause recv() to receive EOF in readstring()
	if conn.Connected {
		logging.Info("irc.shutdown(): Disconnected from server.")
		conn.dispatch(&Line{Cmd: "disconnected"})
		conn.Connected = false
		conn.sock.Close()
		conn.cSend <- true
		conn.cLoop <- true
		conn.cPing <- true
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
	if conn.st != nil {
		str += conn.st.String() + "\n"
	}
	return str
}
