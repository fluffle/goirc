# Copyright 2009 The Go Authors. All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.

include $(GOROOT)/src/Make.inc

TARG=irc/server/config
GOFILES=\
	config.go\
    parser.go\

include $(GOROOT)/src/Make.pkg

parser.go: parser.rl
	ragel -Z -G2 -o parser.go parser.rl

