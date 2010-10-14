# Copyright 2009 The Go Authors. All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.

include $(GOROOT)/src/Make.inc

TARG=rbot
GOFILES=\
	rbot.go\

include $(GOROOT)/src/Make.cmd

all: rbot.conf

rbot.conf: rbot.conf.example
	@if [ -f $@ ] ; then \
		echo "rbot.conf exists, but rbot.conf.example is newer." ; \
	else \
		echo cp $< $@ ; \
		cp $< $@ ; \
	fi
