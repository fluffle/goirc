// Some IRC testing code!

package irc

import (
	"bufio";
	"os";
	"net";
	"fmt";
	"strings";
)

// the IRC connection object
type IRCConn struct {
	sock	*bufio.ReadWriter;
	Host	string;
	Me	string;
	Ident	string;
	Name	string;
	con bool;
	reg	bool;
	events	map[string] []func (*IRCConn, *IRCLine);
	chans	map[string] *IRCChan;
	nicks	map[string] *IRCNick;
}

// We'll parse an incoming line into this struct
// raw =~ ":nick!user@host cmd args[] :text"
// src == "nick!user@host"
type IRCLine struct {
	Nick	string;
	User	string;
	Host	string;
	Src	string;
	Cmd	string;
	Args	[]string;
	Text	string;
	Raw		string;
}

// A struct representing an IRC channel
type IRCChan struct {
	Name	string;
	Topic	string;
	Modes	map[string] string;
	Nicks	map[string] *IRCNick;
}

// A struct representing an IRC nick
type IRCNick struct {
	Name	string;
	Chans	map[string] *IRCChan;
}

// construct a new IRC Connection object
func New(nick, user, name string) (conn *IRCConn) {
	conn = &IRCConn{Me: nick, Ident: user, Name: name};
	// allocate meh some memoraaaahh
	conn.nicks = make(map[string] *IRCNick);
	conn.chans = make(map[string] *IRCChan);
	conn.events = make(map[string] []func(*IRCConn, *IRCLine));
	conn.setupEvents();
	return conn
}

// connect the IRC connection object to a host
func (conn *IRCConn) Connect(host, pass string) (err os.Error) {
	if !hasPort(host) {
		host += ":6667";
	}
	sock, err := net.Dial("tcp", "", host);
	if err != nil {
		return err
	}
	conn.sock = bufio.NewReadWriter(bufio.NewReader(sock), bufio.NewWriter(sock));
	conn.con = true;
	conn.Host = host;

	// initial connection set-up
	// verify valid nick/user/name here?
	if pass != "" {
		conn.Pass(pass)
	}
	conn.Nick(conn.Me);
	conn.User(conn.Ident, conn.Name);

	for line, err := conn.recv(); err == nil; line, err = conn.recv() {
		// initial loop to get us to the point where we're connected
		conn.dispatchEvent(line);
		if line.Cmd == "001" {
			break;
		}
	}
	return err;
}

func (conn *IRCConn) RunLoop(c chan os.Error) {
	var err os.Error;
	for line, err := conn.recv(); err == nil; line, err = conn.recv() {
		conn.dispatchEvent(line);
	}
	c <- err;
	return;
}

// copied from http.client for great justice
func hasPort(s string) bool {
	return strings.LastIndex(s, ":") > strings.LastIndex(s, "]")
}

// send \r\n terminated line to peer, propagate errors
func (conn *IRCConn) send(line string) (err os.Error) {
	err = conn.sock.WriteString(line + "\r\n");
	conn.sock.Flush();
	fmt.Println("-> " + line);
	return err
}

// receive one \r\n terminated line from peer and parse it, propagate errors
func (conn *IRCConn) recv() (line *IRCLine, err os.Error) {
	s, err := conn.sock.ReadString('\n');
	if err != nil {
		return line, err
	}
	// chop off \r\n
	s = s[0:len(s)-2];
	fmt.Println("<- " + s);

	line = &IRCLine{Raw: s};
	if s[0] == ':' {
		// remove a source and parse it
		if idx := strings.Index(s, " "); idx != -1 {
			line.Src, s = s[1:idx], s[idx+1:len(s)];
		} else {
			// pretty sure we shouldn't get here ...
			line.Src = s[1:len(s)];
			return line, nil;
		}

		// src can be the hostname of the irc server or a nick!user@host
		line.Host = line.Src;
		nidx, uidx := strings.Index(line.Src, "!"), strings.Index(line.Src, "@");
		if uidx != -1 && nidx != -1 {
			line.Nick = line.Src[0:nidx];
			line.User = line.Src[nidx+1:uidx];
			line.Host = line.Src[uidx+1:len(line.Src)];
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
	return line, nil
}

