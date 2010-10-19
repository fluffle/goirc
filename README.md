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

### Commands

All commands are prefixed with the trigger configured in rbot.conf.

- `tr text`: detect the language of text
- `tr en|ja en|es text`: translate text into Japanese and Spanish
- `flags raylu`: get's raylu's flags
- `flags`: get's the flags of the user executing the command
- `add john t`: gives john the t flag
- `remove john t`: removes the t flag from john
- `remove john`: removes all of john's flags
- `topic text`: sets the topic and basetopic to text
- `topic`: gets the current basetopic
- `appendtopic text`: if the topic does not starts with basetopic, sets the basetopic to the current topic. Makes the topic basetopic+text.
- `say text`: says text to the channel

Commands that don't require access behave the same when sent to a channel the bot is in and when whispered to the bot. Commands that require access are listed above as if they were sent to a channel. When sent as a whisper, the first argument must be a channel name.

### Flags

Access is configured in auth.conf and based on ident and host; nick is ignored. The owner is configured per server and other access is configured per channel. Owners can use any commands.

The following is a description of the commands enabled by each flag:

- `a`: add remove
- `t`: topic appendtopic
- `s`: say

In addition, a user must have _some_ access to use `flags`.

### Miscellaneous

This project was forked from jessta/goirc which is in turn forked from fluffle/goirc. Both of those projects are focused on developing the goirc framework whereas this is focused on developing a bot.
