GoIRC Client Framework
======================

### Acquiring and Building

Pretty simple, really:

	go get github.com/fluffle/goirc/client

There is some example code that demonstrates usage of the library in `client.go`. This will connect to freenode and join `#go-nuts` by default, so be careful ;-)

See `fix/goirc.go` and the README there for a quick way to migrate from the
old `go1` API.

### Using the framework

Synopsis:

	import irc "github.com/fluffle/goirc/client"

	func main() {
		// Creating a simple IRC client is simple.
		c := irc.SimpleClient("nick")

		// Or, create a config and fiddle with it first:
		cfg := irc.NewConfig("nick")
		cfg.SSL = true
		cfg.Server = "irc.freenode.net:7000"
		cfg.NewNick = func(n string) string { return n + "^" }
		c := irc.Client(cfg)

		// Add handlers to do things here!
		// e.g. join a channel on connect.
		c.HandleFunc("connected",
			func(conn *irc.Conn, line *irc.Line) { conn.Join("#channel") })
		// And a signal on disconnect
		quit := make(chan bool)
		c.HandleFunc("disconnected",
			func(conn *irc.Conn, line *irc.Line) { quit <- true })

		// Tell client to connect.
		if err := c.Connect(); err != nil {
			fmt.Printf("Connection error: %s\n", err.String())
		}

		// With a "simple" client, set Server before calling Connect...
		c.Config().Server = "irc.freenode.net"

		// ... or, use ConnectTo instead.
		if err := c.ConnectTo("irc.freenode.net"); err != nil {
			fmt.Printf("Connection error: %s\n", err.String())
		}

		// Wait for disconnect
		<-quit
	}

The test client provides a good (if basic) example of how to use the framework.
Reading `client/handlers.go` gives a more in-depth look at how handlers can be
written. Commands to be sent to the server (e.g. PRIVMSG) are methods of the
main `*Conn` struct, and can be found in `client/commands.go` (not all of the
possible IRC commands are implemented yet). Events are produced directly from
the messages from the IRC server, so you have to handle e.g. "332" for
`RPL_TOPIC` to get the topic for a channel.

The vast majority of handlers implemented within the framework deal with state
tracking of all nicks in any channels that the client is also present in. These
handers are in `client/state_handlers.go`. State tracking is optional, disabled
by default, and can be enabled and disabled by calling `EnableStateTracking()`
and `DisableStateTracking()` respectively. Doing this while connected to an IRC
server will probably result in an inconsistent state and a lot of warnings to
STDERR ;-)

### Misc.

Sorry the documentation is crap. Use the source, Luke.

[Feedback](mailto:a.bramley@gmail.com) on design decisions is welcome. I am
indebted to Matt Gruen for his work on
[go-bot](http://code.google.com/p/go-bot/source/browse/irc.go) which inspired
the re-organisation and channel-based communication structure of `*Conn.send()`
and `*Conn.recv()`. I'm sure things could be more asynchronous, still.

This code is (c) 2009-13 Alex Bramley, and released under the same licence terms
as Go itself.

Contributions gratefully received from:

  - [iopred](http://github.com/iopred)
  - [StalkR](http://github.com/StalkR)

