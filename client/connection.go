package client

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/fluffle/goirc/logging"
	"github.com/fluffle/goirc/state"
	"golang.org/x/net/proxy"
)

// Conn encapsulates a connection to a single IRC server. Create
// one with Client or SimpleClient.
type Conn struct {
	// For preventing races on (dis)connect.
	mu sync.RWMutex

	// Contains parameters that people can tweak to change client behaviour.
	cfg *Config

	// Handlers
	intHandlers *hSet
	fgHandlers  *hSet
	bgHandlers  *hSet

	// State tracker for nicks and channels
	st         state.Tracker
	stRemovers []Remover

	// I/O stuff to server
	dialer      *net.Dialer
	proxyDialer proxy.Dialer
	sock        net.Conn
	io          *bufio.ReadWriter
	in          chan *Line
	out         chan string
	connected   bool

	// Control channel and WaitGroup for goroutines
	die chan struct{}
	wg  sync.WaitGroup

	// Internal counters for flood protection
	badness  time.Duration
	lastsent time.Time
}

// Config contains options that can be passed to Client to change the
// behaviour of the library during use. It is recommended that NewConfig
// is used to create this struct rather than instantiating one directly.
// Passing a Config with no Nick in the Me field to Client will result
// in unflattering consequences.
type Config struct {
	// Set this to provide the Nick, Ident and Name for the client to use.
	// It is recommended to call Conn.Me to get up-to-date information
	// about the current state of the client's IRC nick after connecting.
	Me *state.Nick

	// Hostname to connect to and optional connect password.
	// Changing these after connection will have no effect until the
	// client reconnects.
	Server, Pass string

	// Are we connecting via SSL? Do we care about certificate validity?
	// Changing these after connection will have no effect until the
	// client reconnects.
	SSL       bool
	SSLConfig *tls.Config

	// To connect via proxy set the proxy url here.
	// Changing these after connection will have no effect until the
	// client reconnects.
	Proxy string

	// Local address to bind to when connecting to the server.
	LocalAddr string

	// To attempt RFC6555 parallel IPv4 and IPv6 connections if both
	// address families are returned for a hostname, set this to true.
	// Passed through to https://golang.org/pkg/net/#Dialer
	DualStack bool

	// Replaceable function to customise the 433 handler's new nick.
	// By default an underscore "_" is appended to the current nick.
	NewNick func(string) string

	// Client->server ping frequency, in seconds. Defaults to 3m.
	// Set to 0 to disable client-side pings.
	PingFreq time.Duration

	// The duration before a connection timeout is triggered. Defaults to 1m.
	// Set to 0 to wait indefinitely.
	Timeout time.Duration

	// Set this to true to disable flood protection and false to re-enable.
	Flood bool

	// Sent as the reply to a CTCP VERSION message.
	Version string

	// Sent as the default QUIT message if Quit is called with no args.
	QuitMessage string

	// Configurable panic recovery for all handlers.
	// Defaults to logging an error, see LogPanic.
	Recover func(*Conn, *Line)

	// Split PRIVMSGs, NOTICEs and CTCPs longer than SplitLen characters
	// over multiple lines. Default to 450 if not set.
	SplitLen int
}

// NewConfig creates a Config struct containing sensible defaults.
// It takes one required argument: the nick to use for the client.
// Subsequent string arguments set the client's ident and "real"
// name, but these are optional.
func NewConfig(nick string, args ...string) *Config {
	cfg := &Config{
		Me:       &state.Nick{Nick: nick},
		PingFreq: 3 * time.Minute,
		NewNick:  func(s string) string { return s + "_" },
		Recover:  (*Conn).LogPanic, // in dispatch.go
		SplitLen: defaultSplit,
		Timeout:  60 * time.Second,
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

// SimpleClient creates a new Conn, passing its arguments to NewConfig.
// If you don't need to change any client options and just want to get
// started quickly, this is a convenient shortcut.
func SimpleClient(nick string, args ...string) *Conn {
	conn := Client(NewConfig(nick, args...))
	return conn
}

// Client takes a Config struct and returns a new Conn ready to have
// handlers added and connect to a server.
func Client(cfg *Config) *Conn {
	if cfg == nil {
		cfg = NewConfig("__idiot__")
	}
	if cfg.Me == nil || cfg.Me.Nick == "" || cfg.Me.Ident == "" {
		cfg.Me = &state.Nick{Nick: "__idiot__"}
		cfg.Me.Ident = "goirc"
		cfg.Me.Name = "Powered by GoIRC"
	}

	dialer := new(net.Dialer)
	dialer.Timeout = cfg.Timeout
	dialer.DualStack = cfg.DualStack
	if cfg.LocalAddr != "" {
		if !hasPort(cfg.LocalAddr) {
			cfg.LocalAddr += ":0"
		}

		local, err := net.ResolveTCPAddr("tcp", cfg.LocalAddr)
		if err == nil {
			dialer.LocalAddr = local
		} else {
			logging.Error("irc.Client(): Cannot resolve local address %s: %s", cfg.LocalAddr, err)
		}
	}

	conn := &Conn{
		cfg:         cfg,
		dialer:      dialer,
		intHandlers: handlerSet(),
		fgHandlers:  handlerSet(),
		bgHandlers:  handlerSet(),
		stRemovers:  make([]Remover, 0, len(stHandlers)),
		lastsent:    time.Now(),
	}
	conn.addIntHandlers()
	return conn
}

// Connected returns true if the client is successfully connected to
// an IRC server. It becomes true when the TCP connection is established,
// and false again when the connection is closed.
func (conn *Conn) Connected() bool {
	conn.mu.RLock()
	defer conn.mu.RUnlock()
	return conn.connected
}

// Config returns a pointer to the Config struct used by the client.
// Many of the elements of Config may be changed at any point to
// affect client behaviour. To disable flood protection temporarily,
// for example, a handler could do:
//
//     conn.Config().Flood = true
//     // Send many lines to the IRC server, risking "excess flood"
//     conn.Config().Flood = false
//
func (conn *Conn) Config() *Config {
	return conn.cfg
}

// Me returns a state.Nick that reflects the client's IRC nick at the
// time it is called. If state tracking is enabled, this comes from
// the tracker, otherwise it is equivalent to conn.cfg.Me.
func (conn *Conn) Me() *state.Nick {
	if conn.st != nil {
		conn.cfg.Me = conn.st.Me()
	}
	return conn.cfg.Me
}

// StateTracker returns the state tracker being used by the client,
// if tracking is enabled, and nil otherwise.
func (conn *Conn) StateTracker() state.Tracker {
	return conn.st
}

// EnableStateTracking causes the client to track information about
// all channels it is joined to, and all the nicks in those channels.
// This can be rather handy for a number of bot-writing tasks. See
// the state package for more details.
//
// NOTE: Calling this while connected to an IRC server may cause the
// state tracker to become very confused all over STDERR if logging
// is enabled. State tracking should enabled before connecting or
// at a pinch while the client is not joined to any channels.
func (conn *Conn) EnableStateTracking() {
	conn.mu.Lock()
	defer conn.mu.Unlock()
	if conn.st == nil {
		n := conn.cfg.Me
		conn.st = state.NewTracker(n.Nick)
		conn.st.NickInfo(n.Nick, n.Ident, n.Host, n.Name)
		conn.cfg.Me = conn.st.Me()
		conn.addSTHandlers()
	}
}

// DisableStateTracking causes the client to stop tracking information
// about the channels and nicks it knows of. It will also wipe current
// state from the state tracker.
func (conn *Conn) DisableStateTracking() {
	conn.mu.Lock()
	defer conn.mu.Unlock()
	if conn.st != nil {
		conn.cfg.Me = conn.st.Me()
		conn.delSTHandlers()
		conn.st.Wipe()
		conn.st = nil
	}
}

// Per-connection state initialisation.
func (conn *Conn) initialise() {
	conn.io = nil
	conn.sock = nil
	conn.in = make(chan *Line, 32)
	conn.out = make(chan string, 32)
	conn.die = make(chan struct{})
	if conn.st != nil {
		conn.st.Wipe()
	}
}

// ConnectTo connects the IRC client to "host[:port]", which should be either
// a hostname or an IP address, with an optional port. It sets the client's
// Config.Server to host, Config.Pass to pass if one is provided, and then
// calls Connect.
func (conn *Conn) ConnectTo(host string, pass ...string) error {
	conn.cfg.Server = host
	if len(pass) > 0 {
		conn.cfg.Pass = pass[0]
	}
	return conn.Connect()
}

// Connect connects the IRC client to the server configured in Config.Server.
// To enable explicit SSL on the connection to the IRC server, set Config.SSL
// to true before calling Connect(). The port will default to 6697 if SSL is
// enabled, and 6667 otherwise.
// To enable connecting via a proxy server, set Config.Proxy to the proxy URL
// (example socks5://localhost:9000) before calling Connect().
//
// Upon successful connection, Connected will return true and a REGISTER event
// will be fired. This is mostly for internal use; it is suggested that a
// handler for the CONNECTED event is used to perform any initial client work
// like joining channels and sending messages.
func (conn *Conn) Connect() error {
	// We don't want to hold conn.mu while firing the REGISTER event,
	// and it's much easier and less error prone to defer the unlock,
	// so the connect mechanics have been delegated to internalConnect.
	err := conn.internalConnect()
	if err == nil {
		conn.dispatch(&Line{Cmd: REGISTER, Time: time.Now()})
	}
	return err
}

// internalConnect handles the work of actually connecting to the server.
func (conn *Conn) internalConnect() error {
	conn.mu.Lock()
	defer conn.mu.Unlock()
	conn.initialise()

	if conn.cfg.Server == "" {
		return fmt.Errorf("irc.Connect(): cfg.Server must be non-empty")
	}
	if conn.connected {
		return fmt.Errorf("irc.Connect(): Cannot connect to %s, already connected.", conn.cfg.Server)
	}

	if !hasPort(conn.cfg.Server) {
		if conn.cfg.SSL {
			conn.cfg.Server = net.JoinHostPort(conn.cfg.Server, "6697")
		} else {
			conn.cfg.Server = net.JoinHostPort(conn.cfg.Server, "6667")
		}
	}

	if conn.cfg.Proxy != "" {
		proxyURL, err := url.Parse(conn.cfg.Proxy)
		if err != nil {
			return err
		}
		conn.proxyDialer, err = proxy.FromURL(proxyURL, conn.dialer)
		if err != nil {
			return err
		}

		logging.Info("irc.Connect(): Connecting to %s.", conn.cfg.Server)
		if s, err := conn.proxyDialer.Dial("tcp", conn.cfg.Server); err == nil {
			conn.sock = s
		} else {
			return err
		}
	} else {
		logging.Info("irc.Connect(): Connecting to %s.", conn.cfg.Server)
		if s, err := conn.dialer.Dial("tcp", conn.cfg.Server); err == nil {
			conn.sock = s
		} else {
			return err
		}
	}

	if conn.cfg.SSL {
		logging.Info("irc.Connect(): Performing SSL handshake.")
		s := tls.Client(conn.sock, conn.cfg.SSLConfig)
		if err := s.Handshake(); err != nil {
			return err
		}
		conn.sock = s
	}

	conn.postConnect(true)
	conn.connected = true
	return nil
}

// postConnect performs post-connection setup, for ease of testing.
func (conn *Conn) postConnect(start bool) {
	conn.io = bufio.NewReadWriter(
		bufio.NewReader(conn.sock),
		bufio.NewWriter(conn.sock))
	if start {
		conn.wg.Add(3)
		go conn.send()
		go conn.recv()
		go conn.runLoop()
		if conn.cfg.PingFreq > 0 {
			conn.wg.Add(1)
			go conn.ping()
		}
	}
}

// hasPort returns true if the string hostname has a :port suffix.
// It was copied from net/http for great justice.
func hasPort(s string) bool {
	return strings.LastIndex(s, ":") > strings.LastIndex(s, "]")
}

// send is started as a goroutine after a connection is established.
// It shuttles data from the output channel to write(), and is killed
// when Conn.die is closed.
func (conn *Conn) send() {
	for {
		select {
		case line := <-conn.out:
			if err := conn.write(line); err != nil {
				logging.Error("irc.send(): %s", err.Error())
				// We can't defer this, because Close() waits for it.
				conn.wg.Done()
				conn.Close()
				return
			}
		case <-conn.die:
			// control channel closed, bail out
			conn.wg.Done()
			return
		}
	}
}

// recv is started as a goroutine after a connection is established.
// It receives "\r\n" terminated lines from the server, parses them into
// Lines, and sends them to the input channel.
func (conn *Conn) recv() {
	for {
		s, err := conn.io.ReadString('\n')
		if err != nil {
			if err != io.EOF {
				logging.Error("irc.recv(): %s", err.Error())
			}
			// We can't defer this, because Close() waits for it.
			conn.wg.Done()
			conn.Close()
			return
		}
		s = strings.Trim(s, "\r\n")
		logging.Debug("<- %s", s)

		if line := ParseLine(s); line != nil {
			line.Time = time.Now()
			conn.in <- line
		} else {
			logging.Warn("irc.recv(): problems parsing line:\n  %s", s)
		}
	}
}

// ping is started as a goroutine after a connection is established, as
// long as Config.PingFreq >0. It pings the server every PingFreq seconds.
func (conn *Conn) ping() {
	defer conn.wg.Done()
	tick := time.NewTicker(conn.cfg.PingFreq)
	for {
		select {
		case <-tick.C:
			conn.Ping(fmt.Sprintf("%d", time.Now().UnixNano()))
		case <-conn.die:
			// control channel closed, bail out
			tick.Stop()
			return
		}
	}
}

// runLoop is started as a goroutine after a connection is established.
// It pulls Lines from the input channel and dispatches them to any
// handlers that have been registered for that IRC verb.
func (conn *Conn) runLoop() {
	defer conn.wg.Done()
	for {
		select {
		case line := <-conn.in:
			conn.dispatch(line)
		case <-conn.die:
			// control channel closed, bail out
			return
		}
	}
}

// write writes a \r\n terminated line of output to the connected server,
// using Hybrid's algorithm to rate limit if conn.cfg.Flood is false.
func (conn *Conn) write(line string) error {
	if !conn.cfg.Flood {
		if t := conn.rateLimit(len(line)); t != 0 {
			// sleep for the current line's time value before sending it
			logging.Info("irc.rateLimit(): Flood! Sleeping for %.2f secs.",
				t.Seconds())
			<-time.After(t)
		}
	}

	if _, err := conn.io.WriteString(line + "\r\n"); err != nil {
		return err
	}
	if err := conn.io.Flush(); err != nil {
		return err
	}
	if strings.HasPrefix(line, "PASS") {
		line = "PASS **************"
	}
	logging.Debug("-> %s", line)
	return nil
}

// rateLimit implements Hybrid's flood control algorithm for outgoing lines.
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

// Close tears down all connection-related state. It is called when either
// the sending or receiving goroutines encounter an error.
// It may also be used to forcibly shut down the connection to the server.
func (conn *Conn) Close() error {
	// Guard against double-call of Close() if we get an error in send()
	// as calling sock.Close() will cause recv() to receive EOF in readstring()
	conn.mu.Lock()
	if !conn.connected {
		conn.mu.Unlock()
		return nil
	}
	logging.Info("irc.Close(): Disconnected from server.")
	conn.connected = false
	err := conn.sock.Close()
	close(conn.die)
	// Drain both in and out channels to avoid a deadlock if the buffers
	// have filled. See TestSendDeadlockOnFullBuffer in connection_test.go.
	conn.drainIn()
	conn.drainOut()
	conn.wg.Wait()
	conn.mu.Unlock()
	// Dispatch after closing connection but before reinit
	// so event handlers can still access state information.
	conn.dispatch(&Line{Cmd: DISCONNECTED, Time: time.Now()})
	return err
}

// drainIn sends all data buffered in conn.in to /dev/null.
func (conn *Conn) drainIn() {
	for {
		select {
		case <-conn.in:
		default:
			return
		}
	}
}

// drainOut does the same for conn.out. Generics!
func (conn *Conn) drainOut() {
	for {
		select {
		case <-conn.out:
		default:
			return
		}
	}
}

// Dumps a load of information about the current state of the connection to a
// string for debugging state tracking and other such things.
func (conn *Conn) String() string {
	str := "GoIRC Connection\n"
	str += "----------------\n\n"
	if conn.Connected() {
		str += "Connected to " + conn.cfg.Server + "\n\n"
	} else {
		str += "Not currently connected!\n\n"
	}
	str += conn.Me().String() + "\n"
	if conn.st != nil {
		str += conn.st.String() + "\n"
	}
	return str
}
