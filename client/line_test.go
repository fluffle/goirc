package client

import (
	"reflect"
	"testing"
	"time"
)

func TestLineCopy(t *testing.T) {
	l1 := &Line{
		Tags:  map[string]string{"foo": "bar", "fizz": "buzz"},
		Nick:  "nick",
		Ident: "ident",
		Host:  "host",
		Src:   "src",
		Cmd:   "cmd",
		Raw:   "raw",
		Args:  []string{"arg", "text"},
		Time:  time.Now(),
	}

	l2 := l1.Copy()

	// Ugly. Couldn't be bothered to bust out reflect and actually think.
	if l2.Tags == nil || l2.Tags["foo"] != "bar" || l2.Tags["fizz"] != "buzz" ||
		l2.Nick != "nick" || l2.Ident != "ident" || l2.Host != "host" ||
		l2.Src != "src" || l2.Cmd != "cmd" || l2.Raw != "raw" ||
		l2.Args[0] != "arg" || l2.Args[1] != "text" || l2.Time != l1.Time {
		t.Errorf("Line not copied correctly")
		t.Errorf("l1: %#v\nl2: %#v", l1, l2)
	}

	// Now, modify l2 and verify l1 not changed
	l2.Tags["foo"] = "baz"
	l2.Nick = l2.Nick[1:]
	l2.Ident = "foo"
	l2.Host = ""
	l2.Args[0] = l2.Args[0][1:]
	l2.Args[1] = "bar"
	l2.Time = time.Now()

	if l2.Tags == nil || l2.Tags["foo"] != "baz" || l2.Tags["fizz"] != "buzz" ||
		l1.Nick != "nick" || l1.Ident != "ident" || l1.Host != "host" ||
		l1.Src != "src" || l1.Cmd != "cmd" || l1.Raw != "raw" ||
		l1.Args[0] != "arg" || l1.Args[1] != "text" || l1.Time == l2.Time {
		t.Errorf("Original modified when copy changed")
		t.Errorf("l1: %#v\nl2: %#v", l1, l2)
	}
}

func TestLineText(t *testing.T) {
	tests := []struct {
		in  *Line
		out string
	}{
		{&Line{}, ""},
		{&Line{Args: []string{"one thing"}}, "one thing"},
		{&Line{Args: []string{"one", "two"}}, "two"},
	}

	for i, test := range tests {
		out := test.in.Text()
		if out != test.out {
			t.Errorf("test %d: expected: '%s', got '%s'", i, test.out, out)
		}
	}
}

func TestLineTarget(t *testing.T) {
	tests := []struct {
		in  *Line
		out string
	}{
		{&Line{}, ""},
		{&Line{Cmd: JOIN, Args: []string{"#foo"}}, "#foo"},
		{&Line{Cmd: PART, Args: []string{"#foo", "bye"}}, "#foo"},
		{&Line{Cmd: PRIVMSG, Args: []string{"Me", "la"}, Nick: "Them"}, "Them"},
		{&Line{Cmd: NOTICE, Args: []string{"Me", "la"}, Nick: "Them"}, "Them"},
		{&Line{Cmd: ACTION, Args: []string{"Me", "la"}, Nick: "Them"}, "Them"},
		{&Line{Cmd: CTCP, Args: []string{"PING", "Me", "1"}, Nick: "Them"}, "Them"},
		{&Line{Cmd: CTCPREPLY, Args: []string{"PONG", "Me", "2"}, Nick: "Them"}, "Them"},
		{&Line{Cmd: PRIVMSG, Args: []string{"#foo", "la"}, Nick: "Them"}, "#foo"},
		{&Line{Cmd: NOTICE, Args: []string{"&foo", "la"}, Nick: "Them"}, "&foo"},
		{&Line{Cmd: ACTION, Args: []string{"!foo", "la"}, Nick: "Them"}, "!foo"},
		{&Line{Cmd: CTCP, Args: []string{"PING", "#foo", "1"}, Nick: "Them"}, "#foo"},
		{&Line{Cmd: CTCPREPLY, Args: []string{"PONG", "#foo", "2"}, Nick: "Them"}, "#foo"},
	}

	for i, test := range tests {
		out := test.in.Target()
		if out != test.out {
			t.Errorf("test %d: expected: '%s', got '%s'", i, test.out, out)
		}
	}
}

func TestLineTags(t *testing.T) {
	tests := []struct {
		in  string
		out *Line
	}{
		{ // Make sure ERROR lines work
			"ERROR :Closing Link: example.org (Too many user connections (global))",
			&Line{
				Nick:  "",
				Ident: "",
				Host:  "",
				Src:   "",
				Cmd:   ERROR,
				Raw:   "ERROR :Closing Link: example.org (Too many user connections (global))",
				Args:  []string{"Closing Link: example.org (Too many user connections (global))"},
			},
		},
		{ // Make sure non-tagged lines work
			":nick!ident@host.com PRIVMSG me :Hello",
			&Line{
				Nick:  "nick",
				Ident: "ident",
				Host:  "host.com",
				Src:   "nick!ident@host.com",
				Cmd:   PRIVMSG,
				Raw:   ":nick!ident@host.com PRIVMSG me :Hello",
				Args:  []string{"me", "Hello"},
			},
		},
		{ // Tags example from the spec
			"@aaa=bbb;ccc;example.com/ddd=eee :nick!ident@host.com PRIVMSG me :Hello",
			&Line{
				Tags:  map[string]string{"aaa": "bbb", "ccc": "", "example.com/ddd": "eee"},
				Nick:  "nick",
				Ident: "ident",
				Host:  "host.com",
				Src:   "nick!ident@host.com",
				Cmd:   PRIVMSG,
				Raw:   "@aaa=bbb;ccc;example.com/ddd=eee :nick!ident@host.com PRIVMSG me :Hello",
				Args:  []string{"me", "Hello"},
			},
		},
		{ // Test escaped characters
			"@\\:=\\:;\\s=\\s;\\r=\\r;\\n=\\n :nick!ident@host.com PRIVMSG me :Hello",
			&Line{
				Tags:  map[string]string{";": ";", " ": " ", "\r": "\r", "\n": "\n"},
				Nick:  "nick",
				Ident: "ident",
				Host:  "host.com",
				Src:   "nick!ident@host.com",
				Cmd:   PRIVMSG,
				Raw:   "@\\:=\\:;\\s=\\s;\\r=\\r;\\n=\\n :nick!ident@host.com PRIVMSG me :Hello",
				Args:  []string{"me", "Hello"},
			},
		},
		{ // Skip empty tag
			"@a=a; :nick!ident@host.com PRIVMSG me :Hello",
			&Line{
				Tags:  map[string]string{"a": "a"},
				Nick:  "nick",
				Ident: "ident",
				Host:  "host.com",
				Src:   "nick!ident@host.com",
				Cmd:   PRIVMSG,
				Raw:   "@a=a; :nick!ident@host.com PRIVMSG me :Hello",
				Args:  []string{"me", "Hello"},
			},
		},
		{ // = in tag
			"@a=a=a; :nick!ident@host.com PRIVMSG me :Hello",
			&Line{
				Tags:  map[string]string{"a": "a=a"},
				Nick:  "nick",
				Ident: "ident",
				Host:  "host.com",
				Src:   "nick!ident@host.com",
				Cmd:   PRIVMSG,
				Raw:   "@a=a=a; :nick!ident@host.com PRIVMSG me :Hello",
				Args:  []string{"me", "Hello"},
			},
		},
	}

	for i, test := range tests {
		got := ParseLine(test.in)
		if !reflect.DeepEqual(got, test.out) {
			t.Errorf("test %d:\nexpected %#v\ngot %#v", i, test.out, got)
		}
	}
}
