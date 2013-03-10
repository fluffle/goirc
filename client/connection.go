package client

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"github.com/fluffle/goirc/state"
	"github.com/fluffle/golog/logging"
	"net"
	"strings"
	"sync"
	"time"
)

// An IRC connection is represented by this struct.
type Conn struct {
	// For preventing races on (dis)connect.
	mu sync.Mutex

	// Contains parameters that people can tweak to change client behaviour.
	cfg *Config

	// Handlers
	handlers *hSet

	// State tracker for nicks and channels
	st         state.Tracker
	stRemovers []Remover

	// I/O stuff to server
	sock      net.Conn
	io        *bufio.ReadWriter
	in        chan *Line
	out       chan string
	connected bool

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

	// Hostname to connect to and optional connect password.
	Server, Pass string

	// Are we connecting via SSL? Do we care about certificate validity?
	SSL       bool
	SSLConfig *tls.Config

	// Replaceable function to customise the 433 handler's new nick
	NewNick func(string) string

	// Client->server ping frequency, in seconds. Defaults to 3m.
	PingFreq time.Duration

	// Set this to true to disable flood protection and false to re-enable
	Flood bool

	// Sent as the reply to a CTCP VERSION message
	Version string

	// Sent as the QUIT message.
	QuitMessage string

	// Configurable panic recovery for all handlers.
	Recover func(*Conn, *Line)
}

func NewConfig(nick string, args ...string) *Config {
	cfg := &Config{
		Me:       state.NewNick(nick),
		PingFreq: 3 * time.Minute,
		NewNick:  func(s string) string { return s + "_" },
		Recover:  (*Conn).LogPanic, // in dispatch.go
	}
	cfg.Me.Ident = "goirc"
	if len(args) > 0 && args[0] != "" {
		cfg.Me.Ident = args[0]
	}
	cfg.Me.Name = "Powered by GoIRC"
	if len(args) > 1 && args[1] != "" {
		cfg.Me.Name = args[1]
	}
	cfg.Version = "Powered by GoIRC"
	cfg.QuitMessage = "GoBye!"
	return cfg
}

// Creates a new IRC connection object, but doesn't connect to anything so
// that you can add event handlers to it. See AddHandler() for details
func SimpleClient(nick string, args ...string) *Conn {
	conn, _ := Client(NewConfig(nick, args...))
	return conn
}

func Client(cfg *Config) (*Conn, error) {
	logging.InitFromFlags()
	if cfg.Me == nil || cfg.Me.Nick == "" || cfg.Me.Ident == "" {
		return nil, fmt.Errorf("irc.Client(): Both cfg.Nick and cfg.Ident must be non-empty.")
	}
	conn := &Conn{
		cfg:        cfg,
		in:         make(chan *Line, 32),
		out:        make(chan string, 32),
		cSend:      make(chan bool),
		cLoop:      make(chan bool),
		cPing:      make(chan bool),
		handlers:   handlerSet(),
		stRemovers: make([]Remover, 0, len(stHandlers)),
		lastsent:   time.Now(),
	}
	conn.addIntHandlers()
	conn.initialise()
	return conn, nil
}

func (conn *Conn) Connected() bool {
	return conn.connected
}

func (conn *Conn) Config() *Config {
	return conn.cfg
}

func (conn *Conn) Me() *state.Nick {
	return conn.cfg.Me
}

func (conn *Conn) StateTracker() state.Tracker {
	return conn.st
}

func (conn *Conn) EnableStateTracking() {
	if conn.st == nil {
		n := conn.cfg.Me
		conn.st = state.NewTracker(n.Nick)
		conn.cfg.Me = conn.st.Me()
		conn.cfg.Me.Ident = n.Ident
		conn.cfg.Me.Name = n.Name
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
func (conn *Conn) ConnectTo(host string, pass ...string) error {
	conn.cfg.Server = host
	if len(pass) > 0 {
		conn.cfg.Pass = pass[0]
	}
	return conn.Connect()
}

func (conn *Conn) Connect() error {
	conn.mu.Lock()
	defer conn.mu.Unlock()

	if conn.cfg.Server == "" {
		return fmt.Errorf("irc.Connect(): cfg.Server must be non-empty")
	}
	if conn.connected {
		return fmt.Errorf("irc.Connect(): Cannot connect to %s, already connected.", conn.cfg.Server)
	}
	if conn.cfg.SSL {
		if !hasPort(conn.cfg.Server) {
			conn.cfg.Server += ":6697"
		}
		logging.Info("irc.Connect(): Connecting to %s with SSL.", conn.cfg.Server)
		if s, err := tls.Dial("tcp", conn.cfg.Server, conn.cfg.SSLConfig); err == nil {
			conn.sock = s
		} else {
			return err
		}
	} else {
		if !hasPort(conn.cfg.Server) {
			conn.cfg.Server += ":6667"
		}
		logging.Info("irc.Connect(): Connecting to %s without SSL.", conn.cfg.Server)
		if s, err := net.Dial("tcp", conn.cfg.Server); err == nil {
			conn.sock = s
		} else {
			return err
		}
	}
	conn.connected = true
	conn.postConnect()
	conn.dispatch(&Line{Cmd: REGISTER})
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
			conn.Ping(fmt.Sprintf("%d", time.Now().UnixNano()))
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
	conn.mu.Lock()
	defer conn.mu.Unlock()
	if !conn.connected {
		return
	}
	logging.Info("irc.shutdown(): Disconnected from server.")
	conn.dispatch(&Line{Cmd: DISCONNECTED})
	conn.connected = false
	conn.sock.Close()
	conn.cSend <- true
	conn.cLoop <- true
	conn.cPing <- true
	// reinit datastructures ready for next connection
	// do this here rather than after runLoop()'s for due to race
	conn.initialise()
}

// Dumps a load of information about the current state of the connection to a
// string for debugging state tracking and other such things.
func (conn *Conn) String() string {
	str := "GoIRC Connection\n"
	str += "----------------\n\n"
	if conn.connected {
		str += "Connected to " + conn.cfg.Server + "\n\n"
	} else {
		str += "Not currently connected!\n\n"
	}
	str += conn.cfg.Me.String() + "\n"
	if conn.st != nil {
		str += conn.st.String() + "\n"
	}
	return str
}
