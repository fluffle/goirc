rbot
======================

### Getting started

Assuming you have go set up (http://golang.org/),

	git clone git://github.com/kless/goconfig.git
	cd goconfig
	make -C config install
	cd ..
	git clone git://github.com/raylu/rbot.git
	cd rbot
	make -C irc install
	make

rbot.conf and auth.conf will be copied. Configure those and then run the bot:

	./rbot

### Misc.

This project was forked from jessta/goirc which is in turn forked from fluffle/goirc. Both of those projects are focused on developing the goirc framework whereas this is focused on developing a bot.
