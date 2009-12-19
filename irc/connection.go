package irc

import (
	"bufio"
	"os"
	"net"
	"fmt"
	"strings"
	"time"
)

// An IRC connection is represented by this struct. Once connected, any errors
// encountered are piped down *Conn.Err; this channel is closed on disconnect.
type Conn struct {
	// Connection Hostname and Nickname
	Host string
	Me   *Nick

	// I/O stuff to server
	sock      *net.TCPConn
	io        *bufio.ReadWriter
	in        chan *Line
	out       chan string
	connected bool

	// Error channel to transmit any fail back to the user
	Err chan os.Error

	// Set this to true before connect to disable throttling
	Flood bool;

	// Event handler mapping
	events map[string][]func(*Conn, *Line)

	// Map of channels we're on
	chans map[string]*Channel

	// Map of nicks we know about
	nicks map[string]*Nick
}

// We parse an incoming line into this struct. Line.Cmd is used as the trigger
// name for incoming event handlers, see *Conn.recv() for details.
//   Raw =~ ":nick!user@host cmd args[] :text"
//   Src == "nick!user@host"
//   Cmd == e.g. PRIVMSG, 332
type Line struct {
	Nick, Ident, Host, Src string
	Cmd, Text, Raw         string
	Args                   []string
}

// Creates a new IRC connection object, but doesn't connect to anything so
// that you can add event handlers to it. See AddHandler() for details.
func New(nick, user, name string) *Conn {
	conn := new(Conn)
	conn.initialise()
	conn.Me = conn.NewNick(nick, user, name, "")
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
}

// Connect the IRC connection object to "host[:port]" which should be either
// a hostname or an IP address, with an optional port defaulting to 6667.
// You can also provide an optional connect password.
func (conn *Conn) Connect(host string, pass ...) os.Error {
	if conn.connected {
		return os.NewError(fmt.Sprintf("irc.Connect(): already connected to %s, cannot connect to %s", conn.Host, host))
	}
	if !hasPort(host) {
		host += ":6667"
	}

	if addr, err := net.ResolveTCPAddr(host); err != nil {
		return err
	} else if conn.sock, err = net.DialTCP("tcp", nil, addr); err != nil {
		return err
	}
	conn.Host = host

	conn.io = bufio.NewReadWriter(
		bufio.NewReader(conn.sock),
		bufio.NewWriter(conn.sock))
	go conn.send()
	go conn.recv()

	// see getStringMsg() in commands.go for what this does
	if p := getStringMsg(pass); p != "" {
		conn.Pass(p)
	}
	conn.Nick(conn.Me.Nick)
	conn.User(conn.Me.Ident, conn.Me.Name)

	go conn.runLoop()
	return nil
}

// dispatch a nicely formatted os.Error to the error channel
func (conn *Conn) error(s string, a ...) { conn.Err <- os.NewError(fmt.Sprintf(s, a)) }

// copied from http.client for great justice
func hasPort(s string) bool { return strings.LastIndex(s, ":") > strings.LastIndex(s, "]") }

// dispatch input from channel as \r\n terminated line to peer
// flood controlled using hybrid's algorithm if conn.Flood is true
func (conn *Conn) send() {
	for {
		line := <-conn.out
		if closed(conn.out) {
			break
		}
		if err := conn.io.WriteString(line + "\r\n"); err != nil {
			conn.error("irc.send(): %s", err.String())
			conn.shutdown()
			break
		}
		conn.io.Flush()
		fmt.Println("-> " + line)

		// Current flood-control implementation is naive and may lead to
		// much frustration. Hybrid's flooding algorithm allows one line every
		// two seconds, and a 120-character-per-second penalty on top of this.
		// We currently just sleep for the correct delay after sending the line
		// but if there's a *lot* of flood, conn.out may fill it's buffers and
		// cause other things to hang within runloop :-(
		if !conn.Flood {
			time.Sleep(2*1000000000 + len(line)*8333333)
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
		// chop off \r\n
		s = s[0 : len(s)-2]
		fmt.Println("<- " + s)

		line := &Line{Raw: s}
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
			line.Text = args[1]
		}
		args = strings.Split(args[0], " ", 0)
		line.Cmd = strings.ToUpper(args[0])
		if len(args) > 1 {
			line.Args = args[1:len(args)]
		}
		conn.in <- line
	}
}

func (conn *Conn) runLoop() {
	for {
		if closed(conn.in) {
			break
		}
		select {
		case line := <-conn.in:
			conn.dispatchEvent(line)
		}
	}
	// if we fall off the end here due to shutdown,
	// reinit everything once the runloop is done
	// so that Connect() can be called again.
	conn.initialise()
}

func (conn *Conn) shutdown() {
	close(conn.in)
	close(conn.out)
	close(conn.Err)
	conn.connected = false
	conn.sock.Close()
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
