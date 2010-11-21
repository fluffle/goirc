GoIRC Client Framework
======================

### Acquiring and Building

Pretty simple, really:

	goinstall github.com/fluffle/goirc

You can build the test client also with:

	make
	./gobot

This will connect to freenode and join `#go-nuts` by default, so be careful ;-)

### Using the framework

Synopsis:

    import irc "github.com/fluffle/goirc/client"
    func main() {
        c := irc.New("nick", "ident", "real name")
        // Optionally, turn on debugging
        c.Debug = true
        // Optionally, enable SSL
        c.SSL = true
        // add handlers to do things here!
	    if err := c.Connect("irc.freenode.net"); err != nil {
		    fmt.Printf("Connection error: %s\n", err.String())
	    }
        for {
            if closed(c.Err) {
                break
            }
            if err := <-c.Err; err != nil {
                fmt.Printf("goirc error: %s", err.String())
            }
        }
    }

The test client provides a good (if basic) example of how to use the framework.
Reading `client/handlers.go` gives a more in-depth look at how handlers can be
written. Commands to be sent to the server (e.g. PRIVMSG) are methods of the
main `*Conn` struct, and can be found in `client/commands.go` (not all of the
possible IRC commands are implemented yet). Events are produced directly from
the messages from the IRC server, so you have to handle e.g. "332" for
`RPL_TOPIC` to get the topic for a channel.

The vast majority of handlers implemented within the framework deal with state
tracking of all nicks in any channels that the client is also present in. It's
likely that this state tracking will become optional in the near future.

### Misc.

Sorry the documentation is crap. Use the source, Luke.

[Feedback](mailto:a.bramley@gmail.com) on design decisions is welcome. I am
indebted to Matt Gruen for his work on
[go-bot](http://code.google.com/p/go-bot/source/browse/irc.go) which inspired
the re-organisation and channel-based communication structure of `*Conn.send()`
and `*Conn.recv()`. I'm sure things could be more asynchronous, still.

This code is (c) 2009-10 Alex Bramley, and released under the same licence terms
as Go itself.
