[![Build Status](https://api.travis-ci.org/fluffle/goirc.svg)](https://travis-ci.org/fluffle/goirc)

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
```go
package main

import (
	"crypto/tls"
	"fmt"

	irc "github.com/fluffle/goirc/client"
)

func main() {
	// Creating a simple IRC client is simple.
	c := irc.SimpleClient("nick")

	// Or, create a config and fiddle with it first:
	cfg := irc.NewConfig("nick")
	cfg.SSL = true
	cfg.SSLConfig = &tls.Config{ServerName: "irc.freenode.net"}
	cfg.Server = "irc.freenode.net:7000"
	cfg.NewNick = func(n string) string { return n + "^" }
	c = irc.Client(cfg)

	// Add handlers to do things here!
	// e.g. join a channel on connect.
	c.HandleFunc(irc.CONNECTED,
		func(conn *irc.Conn, line *irc.Line) { conn.Join("#channel") })
	// And a signal on disconnect
	quit := make(chan bool)
	c.HandleFunc(irc.DISCONNECTED,
		func(conn *irc.Conn, line *irc.Line) { quit <- true })

	// Tell client to connect.
	if err := c.Connect(); err != nil {
		fmt.Printf("Connection error: %s\n", err.Error())
	}

	// With a "simple" client, set Server before calling Connect...
	c.Config().Server = "irc.freenode.net"

	// ... or, use ConnectTo instead.
	if err := c.ConnectTo("irc.freenode.net"); err != nil {
		fmt.Printf("Connection error: %s\n", err.Error())
	}

	// Wait for disconnect
	<-quit
}
```

The test client provides a good (if basic) example of how to use the framework.
Reading `client/handlers.go` gives a more in-depth look at how handlers can be
written. Commands to be sent to the server (e.g. PRIVMSG) are methods of the
main `*Conn` struct, and can be found in `client/commands.go` (not all of the
possible IRC commands are implemented yet). Events are produced directly from
the messages from the IRC server, so you have to handle e.g. "332" for
`RPL_TOPIC` to get the topic for a channel.

The vast majority of handlers implemented within the framework deal with state
tracking of all nicks in any channels that the client is also present in. These
handlers are in `client/state_handlers.go`. State tracking is optional, disabled
by default, and can be enabled and disabled by calling `EnableStateTracking()`
and `DisableStateTracking()` respectively. Doing this while connected to an IRC
server will probably result in an inconsistent state and a lot of warnings to
STDERR ;-)

### Projects using GoIRC

- [xdcc-cli](https://github.com/ostafen/xdcc-cli): A command line tool for searching and downloading files from the IRC network.


### Misc.

Sorry the documentation is crap. Use the source, Luke.

[Feedback](mailto:a.bramley@gmail.com) on design decisions is welcome. I am
indebted to Matt Gruen for his work on
[go-bot](http://code.google.com/p/go-bot/source/browse/irc.go) which inspired
the re-organisation and channel-based communication structure of `*Conn.send()`
and `*Conn.recv()`. I'm sure things could be more asynchronous, still.

This code is (c) 2009-23 Alex Bramley, and released under the same licence terms
as Go itself.

Contributions gratefully received from:

  - [3onyc](https://github.com/3onyc)
  - [bramp](https://github.com/bramp)
  - [cgt](https://github.com/cgt)
  - [iopred](https://github.com/iopred)
  - [Krayons](https://github.com/Krayons)
  - [StalkR](https://github.com/StalkR)
  - [sztanpet](https://github.com/sztanpet)
  - [wathiede](https://github.com/wathiede)
  - [scrapbird](https://github.com/scrapbird)
  - [soul9](https://github.com/soul9)
  - [jakebailey](https://github.com/jakebailey)
  - [stapelberg](https://github.com/stapelberg)
  - [shammash](https://github.com/shammash)
  - [ostafen](https://github.com/ostafen)
  - [supertassu](https://github.com/supertassu)

And thanks to the following for minor doc/fix PRs:

  - [tmcarr](https://github.com/tmcarr)
  - [Gentux](https://github.com/Gentux)
  - [kidanger](https://github.com/kidanger)
  - [ripcurld00d](https://github.com/ripcurld00d)
  - [gundalow](https://github.com/gundalow)
