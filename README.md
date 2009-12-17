GoIRC Client Framework
======================

### Acquiring and Building

Pretty simple, really:

	git clone git://github.com/fluffle/goirc.git
	make -C irc install

You can build the test client also with:

	make
	./gobot

This will connect to freenode and join #go-lang by default, so be careful ;-)

### Using the framework

The test client provides a good (if basic) example of how to use the framework.
Reading irc/handlers.go gives a more in-depth look at how handlers can be
written. Commands to be sent to the server (e.g. PRIVMSG) are methods of the
main \*irc.Conn object, and can be found in irc/commands.go (not all of the
possible IRC commands are implemented yet). Events are produced directly from
the messages from the IRC server, so you have to handle e.g. "332" for
RPL\_TOPIC to get the topic for a channel.

The vast majority of handlers implemented within the framework implement state
tracking of all nicks in channels that the client is also present in. It's
likely that this state tracking will become optional in the near future.

### Misc.

Sorry the documentation is crap. Use the source, Luke.

This code is (c) 2009 Alex Bramley, and released under the same licence terms
as Go itself.
