package client

import (
	"bufio"
	"os"
	"net"
	"crypto/tls"
	"fmt"
	"strings"
	"time"
)

// An IRC connection is represented by this struct. Once connected, any errors
// encountered are piped down *Conn.Err; this channel is closed on disconnect.
type Conn struct {
	// Connection Hostname and Nickname
	Host    string
	Me      *Nick
	Network string

	// Event handler mapping
	events map[string][]func(*Conn, *Line)
	// Map of channels we're on
	chans map[string]*Channel
	// Map of nicks we know about
	nicks map[string]*Nick

	// I/O stuff to server
	sock      net.Conn
	io        *bufio.ReadWriter
	in        chan *Line
	out       chan string
	connected bool

	// Error channel to transmit any fail back to the user
	Err chan os.Error

	// Misc knobs to tweak client behaviour:
	// Are we connecting via SSL? Do we care about certificate validity?
	SSL       bool
	SSLConfig *tls.Config

	// Set this to true to disable flood protection and false to re-enable
	Flood bool

	// Function which returns a *time.Time for use as a timestamp
	Timestamp func() *time.Time

	// Enable debugging? Set format for timestamps on debug output.
	Debug    bool
	TSFormat string
}

// We parse an incoming line into this struct. Line.Cmd is used as the trigger
// name for incoming event handlers, see *Conn.recv() for details.
//   Raw =~ ":nick!user@host cmd args[] :text"
//   Src == "nick!user@host"
//   Cmd == e.g. PRIVMSG, 332
type Line struct {
	Nick, Ident, Host, Src string
	Cmd, Raw               string
	Args                   []string
	Time                   *time.Time
}

// Creates a new IRC connection object, but doesn't connect to anything so
// that you can add event handlers to it. See AddHandler() for details.
func New(nick, user, name string) *Conn {
	conn := new(Conn)
	conn.initialise()
	conn.SSL = false
	conn.SSLConfig = nil
	conn.Me = conn.NewNick(nick, user, name, "")
	conn.Timestamp = time.LocalTime
	conn.TSFormat = "15:04:05"
	conn.setupEvents()
	return conn
}

func (conn *Conn) initialise() {
	// allocate meh some memoraaaahh
	conn.nicks = make(map[string]*Nick)
	conn.chans = make(map[string]*Channel)
	conn.in = make(chan *Line, 32)
	conn.out = make(chan string, 32)
	conn.Err = make(chan os.Error, 4)
	conn.io = nil
	conn.sock = nil

	// if this is being called because we are reconnecting, conn.Me
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
		// It's unfortunate that tls.Dial doesn't allow a tls.Config arg,
		// so we simply replicate it here with the correct Config.
		// http://codereview.appspot.com/2883041
		if s, err := net.Dial("tcp", host); err == nil {
			// Passing nil config => certs are validated.
			c := tls.Client(s, conn.SSLConfig)
			if err = c.Handshake(); err == nil {
				conn.sock = c
			} else {
				s.Close()
				return err
			}
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
	conn.io = bufio.NewReadWriter(
		bufio.NewReader(conn.sock),
		bufio.NewWriter(conn.sock))
	conn.sock.SetTimeout(300000000000) // 5 minutes
	go conn.send()
	go conn.recv()

	if len(pass) > 0 {
		conn.Pass(pass[0])
	}
	conn.Nick(conn.Me.Nick)
	conn.User(conn.Me.Ident, conn.Me.Name)

	go conn.runLoop()
	return nil
}

// dispatch a nicely formatted os.Error to the error channel
func (conn *Conn) error(s string, a ...interface{}) {
	conn.Err <- os.NewError(fmt.Sprintf(s, a...))
}

// copied from http.client for great justice
func hasPort(s string) bool {
	return strings.LastIndex(s, ":") > strings.LastIndex(s, "]")
}

// dispatch input from channel as \r\n terminated line to peer
// flood controlled using hybrid's algorithm if conn.Flood is true
func (conn *Conn) send() {
	lastsent := time.Nanoseconds()
	var badness, linetime, second int64 = 0, 0, 1000000000
	for line := range conn.out {
		// Hybrid's algorithm allows for 2 seconds per line and an additional
		// 1/120 of a second per character on that line.
		linetime = 2*second + int64(len(line))*second/120
		if !conn.Flood && conn.connected {
			// No point in tallying up flood protection stuff until connected
			if badness += linetime + lastsent - time.Nanoseconds(); badness < 0 {
				// negative badness times are badness...
				badness = int64(0)
			}
		}
		lastsent = time.Nanoseconds()

		// If we've sent more than 10 second's worth of lines according to the
		// calculation above, then we're at risk of "Excess Flood".
		if badness > 10*second && !conn.Flood {
			// so sleep for the current line's time value before sending it
			time.Sleep(linetime)
		}
		if _, err := conn.io.WriteString(line + "\r\n"); err != nil {
			conn.error("irc.send(): %s", err.String())
			conn.shutdown()
			break
		}
		conn.io.Flush()
		if conn.Debug {
			fmt.Println(conn.Timestamp().Format(conn.TSFormat) + " -> " + line)
		}
	}
}

// receive one \r\n terminated line from peer, parse and dispatch it
func (conn *Conn) recv() {
	for {
		s, err := conn.io.ReadString('\n')
		t := conn.Timestamp()
		if err != nil {
			conn.error("irc.recv(): %s", err.String())
			conn.shutdown()
			break
		}
		s = strings.Trim(s, "\r\n")
		if conn.Debug {
			fmt.Println(t.Format(conn.TSFormat) + " <- " + s)
		}

		line := &Line{Raw: s, Time: t}
		if s[0] == ':' {
			// remove a source and parse it
			if idx := strings.Index(s, " "); idx != -1 {
				line.Src, s = s[1:idx], s[idx+1:len(s)]
			} else {
				// pretty sure we shouldn't get here ...
				line.Src = s[1:len(s)]
				conn.in <- line
				continue
			}

			// src can be the hostname of the irc server or a nick!user@host
			line.Host = line.Src
			nidx, uidx := strings.Index(line.Src, "!"), strings.Index(line.Src, "@")
			if uidx != -1 && nidx != -1 {
				line.Nick = line.Src[0:nidx]
				line.Ident = line.Src[nidx+1 : uidx]
				line.Host = line.Src[uidx+1 : len(line.Src)]
			}
		}

		// now we're here, we've parsed a :nick!user@host or :server off
		// s should contain "cmd args[] :text"
		args := strings.Split(s, " :", 2)
		if len(args) > 1 {
			args = append(strings.Fields(args[0]), args[1])
		} else {
			args = strings.Fields(args[0])
		}
		line.Cmd = strings.ToUpper(args[0])
		if len(args) > 1 {
			line.Args = args[1:len(args)]
		}
		conn.in <- line
	}
}

func (conn *Conn) runLoop() {
	for line := range conn.in {
		conn.dispatchEvent(line)
	}
}

func (conn *Conn) shutdown() {
	close(conn.in)
	close(conn.out)
	close(conn.Err)
	conn.connected = false
	conn.sock.Close()
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
