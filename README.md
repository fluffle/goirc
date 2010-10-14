rbot
======================

### Getting started

Assuming you have go set up (http://golang.org/), first clone http://github.com/kless/goconfig/ and in that directory, run

	make -C config install

Then build the goirc framework in the rbot directory with

	make -C irc install

Finally, build the bot with:

	make

rbot.conf will be copied. Configure that and then run the bot:

	./rbot

### Misc.

This project was forked from jessta/goirc which is in turn forked from fluffle/goirc. Both of those projects are focused on developing the goirc framework whereas this is focused on developing a bot.
