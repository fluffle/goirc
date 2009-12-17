package irc

import (
	"bufio";
	"os";
	"net";
	"fmt";
	"strings";
)

// the IRC connection object
type Conn struct {
	// Hostname, Nickname, etc.
	Host string;
	Me *Nick;

	// I/O stuff to server
	sock	*net.TCPConn;
	io	*bufio.ReadWriter;
	in	chan *Line;
	out	chan string;
	connected bool;

	// Error channel to transmit any fail back to the user
	Err	chan os.Error;

	// Event handler mapping
	events	map[string] []func (*Conn, *Line);

	// Map of channels we're on
	chans	map[string] *Channel;

	// Map of nicks we know about
	nicks	map[string] *Nick;
}

// We'll parse an incoming line into this struct
// raw =~ ":nick!user@host cmd args[] :text"
// src == "nick!user@host"
type Line struct {
	Nick, Ident, Host, Src	string;
	Cmd, Text, Raw	string;
	Args	[]string;
}

// construct a new IRC Connection object
func New(nick, user, name string) *Conn {
	conn := new(Conn);
	conn.initialise();
	conn.Me = conn.NewNick(nick, user, name, "");
	conn.setupEvents();
	return conn;
}

func (conn *Conn) initialise() {
	// allocate meh some memoraaaahh
	fmt.Println("irc.initialise(): initialising...");
	conn.nicks = make(map[string] *Nick);
	conn.chans = make(map[string] *Channel);
	conn.in = make(chan *Line, 32);
	conn.out = make(chan string, 32);
	conn.Err = make(chan os.Error, 4);
	conn.io = nil;
	conn.sock = nil;
}

// connect the IRC connection object to a host
func (conn *Conn) Connect(host, pass string) os.Error {
	if conn.connected {
		return os.NewError(fmt.Sprintf("irc.Connect(): already connected to %s, cannot connect to %s", conn.Host, host));
	}
	if !hasPort(host) {
		host += ":6667";
	}

	if addr, err := net.ResolveTCPAddr(host); err != nil {
		return err
	} else if conn.sock, err = net.DialTCP("tcp", nil, addr); err != nil {
		return err
	}
	fmt.Println("irc.Connect(): connected happily...");
	conn.Host = host;

	conn.io = bufio.NewReadWriter(
		bufio.NewReader(conn.sock),
		bufio.NewWriter(conn.sock),
	);
	go conn.send();
	go conn.recv();

	if pass != "" {
		conn.Pass(pass)
	}
	conn.Nick(conn.Me.Nick);
	conn.User(conn.Me.Ident, conn.Me.Name);

	go conn.runLoop();
	fmt.Println("irc.Connect(): launched runLoop() goroutine.");
	return nil;
}

// dispatch a nicely formatted os.Error to the error channel
func (conn *Conn) error(s string, a ...) {
	conn.Err <- os.NewError(fmt.Sprintf(s, a));
}

// copied from http.client for great justice
func hasPort(s string) bool {
	return strings.LastIndex(s, ":") > strings.LastIndex(s, "]")
}

// dispatch input from channel as \r\n terminated line to peer
func (conn *Conn) send() {
	for {
		line := <-conn.out;
		if closed(conn.out) {
			break;
		}
		if err := conn.io.WriteString(line + "\r\n"); err != nil {
			conn.error("irc.send(): %s", err.String());
			conn.shutdown();
			break;
		}
		conn.io.Flush();
		fmt.Println("-> " + line);
	}
}

// receive one \r\n terminated line from peer, parse and dispatch it
func (conn *Conn) recv() {
	for {
		s, err := conn.io.ReadString('\n');
		if err != nil {
			conn.error("irc.recv(): %s", err.String());
			conn.shutdown();
			break;
		}
		// chop off \r\n
		s = s[0:len(s)-2];
		fmt.Println("<- " + s);

		line := &Line{Raw: s};
		if s[0] == ':' {
			// remove a source and parse it
			if idx := strings.Index(s, " "); idx != -1 {
				line.Src, s = s[1:idx], s[idx+1:len(s)];
			} else {
				// pretty sure we shouldn't get here ...
				line.Src = s[1:len(s)];
				conn.in <- line;
				continue;
			}

			// src can be the hostname of the irc server or a nick!user@host
			line.Host = line.Src;
			nidx, uidx := strings.Index(line.Src, "!"), strings.Index(line.Src, "@");
			if uidx != -1 && nidx != -1 {
				line.Nick  = line.Src[0:nidx];
				line.Ident = line.Src[nidx+1:uidx];
				line.Host  = line.Src[uidx+1:len(line.Src)];
			}
		}

		// now we're here, we've parsed a :nick!user@host or :server off
		// s should contain "cmd args[] :text"
		args := strings.Split(s, " :", 2);
		if len(args) > 1 {
			line.Text = args[1];
		}
		args = strings.Split(args[0], " ", 0);
		line.Cmd = strings.ToUpper(args[0]);
		if len(args) > 1 {
			line.Args = args[1:len(args)];
		}
		conn.in <- line
	}
}

func (conn *Conn) runLoop() {
	for {
		if closed(conn.in) {
			break;
		}
		select {
			case line := <-conn.in:
				conn.dispatchEvent(line);
		}
	}
	fmt.Println("irc.runLoop(): Exited runloop...");
	// if we fall off the end here due to shutdown,
	// reinit everything once the runloop is done
	// so that Connect() can be called again.
	conn.initialise();
}

func (conn *Conn) shutdown() {
	close(conn.in);
	close(conn.out);
	close(conn.Err);
	conn.connected = false;
	conn.sock.Close();
	fmt.Println("irc.shutdown(): shut down sockets and channels!");
}

func (conn *Conn) String() string {
	str := "GoIRC Connection\n";
	str += "----------------\n\n";
	if conn.connected {
		str += "Connected to " + conn.Host + "\n\n"
	} else {
		str += "Not currently connected!\n\n";
	}
	str += conn.Me.String() + "\n";
	str += "GoIRC Channels\n";
	str += "--------------\n\n";
	for _, ch := range conn.chans {
		str += ch.String() + "\n"
	}
	str += "GoIRC NickNames\n";
	str += "---------------\n\n";
	for _, n := range conn.nicks {
		if n != conn.Me {
			str += n.String() + "\n"
		}
	}
	return str;
}

