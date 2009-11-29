package main

import (
	"./irc/_obj/irc";
	"fmt";
	"os";
)

func main() {
	c := irc.New("GoTest", "gotest", "GoBot");
	c.AddHandler("connected",
		func(conn *irc.IRCConn, line *irc.IRCLine) {
			conn.Join("#");
		}
	);
	c.AddHandler("join",
		func(conn *irc.IRCConn, line *irc.IRCLine) {
			if line.Nick == conn.Me {
				conn.Privmsg(line.Text, "I LIVE, BITCHES");
			}
		}
	);
	if err := c.Connect("irc.pl0rt.org", ""); err != nil {
		fmt.Printf("Connection error: %v\n", err);
		return;
	}

	// if we get here, we're successfully connected and should have just
	// dispatched the "CONNECTED" event to it's handlers \o/
	control := make(chan os.Error, 1);
	go c.RunLoop(control);
	if err := <-control; err != nil {
		fmt.Printf("IRCConn.RunLoop terminated: %v\n", err);
	}
}
