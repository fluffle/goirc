GoIRC Client Framework
======================

### Acquiring and Building

	git clone git://github.com/raylu/rbot.git
	make -C irc install

You can build the bot with:

	make
	./gobot

This will connect to rizon and join `#vn-meta` by default, so be careful ;-)

### Using the framework

Synopsis:

    import "irc"
    func main() {
        c := irc.New("nick", "ident", "real name")
        // add handlers to do things here!
	    if err := c.Connect("irc.freenode.net", ""); err != nil {
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
Reading `irc/handlers.go` gives a more in-depth look at how handlers can be
written. Commands to be sent to the server (e.g. PRIVMSG) are methods of the
main `*irc.Conn` object, and can be found in `irc/commands.go` (not all of the
possible IRC commands are implemented yet). Events are produced directly from
the messages from the IRC server, so you have to handle e.g. "332" for
`RPL_TOPIC` to get the topic for a channel.

The vast majority of handlers implemented within the framework deal with state
tracking of all nicks in any channels that the client is also present in. It's
likely that this state tracking will become optional in the near future.

### Misc.

This project was forked from jessta/goirc which is in turn forked from
fluffle/goirc. Both of those projects are focused on developing the goirc
framework whereas this is focused on developing a bot.
