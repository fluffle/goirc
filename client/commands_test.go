package client

import (
	"reflect"
	"testing"
)

func TestCutNewLines(t *testing.T) {
	tests := []struct{ in, out string }{
		{"", ""},
		{"foo bar", "foo bar"},
		{"foo bar\rbaz", "foo bar"},
		{"foo bar\nbaz", "foo bar"},
		{"blorp\r\n\r\nbloop", "blorp"},
		{"\n\rblaap", ""},
		{"\r\n", ""},
		{"boo\\r\\n\\n\r", "boo\\r\\n\\n"},
	}
	for i, test := range tests {
		out := cutNewLines(test.in)
		if test.out != out {
			t.Errorf("test %d: expected %q, got %q", i, test.out, out)
		}
	}
}

func TestIndexFragment(t *testing.T) {
	tests := []struct {
		in  string
		out int
	}{
		{"", -1},
		{"foobarbaz", -1},
		{"foo bar baz", 8},
		{"foo. bar baz", 5},
		{"foo: bar baz", 5},
		{"foo; bar baz", 5},
		{"foo, bar baz", 5},
		{"foo! bar baz", 5},
		{"foo? bar baz", 5},
		{"foo\" bar baz", 5},
		{"foo' bar baz", 5},
		{"foo. bar. baz beep", 10},
		{"foo. bar, baz beep", 10},
	}
	for i, test := range tests {
		out := indexFragment(test.in)
		if test.out != out {
			t.Errorf("test %d: expected %d, got %d", i, test.out, out)
		}
	}
}

func TestSplitMessage(t *testing.T) {
	tests := []struct {
		in  string
		sp  int
		out []string
	}{
		{"", 0, []string{""}},
		{"foo", 0, []string{"foo"}},
		{"foo bar baz beep", 0, []string{"foo bar baz beep"}},
		{"foo bar baz beep", 15, []string{"foo bar baz ...", "beep"}},
		{"foo bar, baz beep", 15, []string{"foo bar, ...", "baz beep"}},
		{"0123456789012345", 0, []string{"0123456789012345"}},
		{"0123456789012345", 15, []string{"012345678901...", "2345"}},
		{"0123456789012345", 16, []string{"0123456789012345"}},
	}
	for i, test := range tests {
		out := splitMessage(test.in, test.sp)
		if !reflect.DeepEqual(test.out, out) {
			t.Errorf("test %d: expected %q, got %q", i, test.out, out)
		}
	}
}

func TestClientCommands(t *testing.T) {
	c, s := setUp(t)
	defer s.tearDown()

	// Avoid having to type ridiculously long lines to test that
	// messages longer than SplitLen are correctly sent to the server.
	c.cfg.SplitLen = 23

	c.Pass("password")
	s.nc.Expect("PASS password")

	c.Nick("test")
	s.nc.Expect("NICK test")

	c.User("test", "Testing IRC")
	s.nc.Expect("USER test 12 * :Testing IRC")

	c.Raw("JUST a raw :line")
	s.nc.Expect("JUST a raw :line")

	c.Join("#foo")
	s.nc.Expect("JOIN #foo")
	c.Join("#foo bar")
	s.nc.Expect("JOIN #foo bar")

	c.Part("#foo")
	s.nc.Expect("PART #foo")
	c.Part("#foo", "Screw you guys...")
	s.nc.Expect("PART #foo :Screw you guys...")

	c.Quit()
	s.nc.Expect("QUIT :GoBye!")
	c.Quit("I'm going home.")
	s.nc.Expect("QUIT :I'm going home.")

	c.Whois("somebody")
	s.nc.Expect("WHOIS somebody")

	c.Who("*@some.host.com")
	s.nc.Expect("WHO *@some.host.com")

	c.Privmsg("#foo", "bar")
	s.nc.Expect("PRIVMSG #foo :bar")

	c.Privmsgln("#foo", "bar")
	s.nc.Expect("PRIVMSG #foo :bar")

	c.Privmsgf("#foo", "say %s", "foo")
	s.nc.Expect("PRIVMSG #foo :say foo")

	c.Privmsgln("#foo", "bar", 1, 3.54, []int{24, 36})
	s.nc.Expect("PRIVMSG #foo :bar 1 3.54 [24 36]")

	c.Privmsgf("#foo", "user %d is at %s", 2, "home")
	s.nc.Expect("PRIVMSG #foo :user 2 is at home")

	//                 0123456789012345678901234567890123
	c.Privmsg("#foo", "foo bar baz blorp. woo woobly woo.")
	s.nc.Expect("PRIVMSG #foo :foo bar baz blorp. ...")
	s.nc.Expect("PRIVMSG #foo :woo woobly woo.")

	c.Privmsgln("#foo", "foo bar baz blorp. woo woobly woo.")
	s.nc.Expect("PRIVMSG #foo :foo bar baz blorp. ...")
	s.nc.Expect("PRIVMSG #foo :woo woobly woo.")

	c.Privmsgf("#foo", "%s %s", "foo bar baz blorp.", "woo woobly woo.")
	s.nc.Expect("PRIVMSG #foo :foo bar baz blorp. ...")
	s.nc.Expect("PRIVMSG #foo :woo woobly woo.")

	c.Privmsgln("#foo", "foo bar", 3.54, "blorp.", "woo", "woobly", []int{1, 2})
	s.nc.Expect("PRIVMSG #foo :foo bar 3.54 blorp. ...")
	s.nc.Expect("PRIVMSG #foo :woo woobly [1 2]")

	c.Privmsgf("#foo", "%s %.2f %s %s %s %v", "foo bar", 3.54, "blorp.", "woo", "woobly", []int{1, 2})
	s.nc.Expect("PRIVMSG #foo :foo bar 3.54 blorp. ...")
	s.nc.Expect("PRIVMSG #foo :woo woobly [1 2]")

	c.Notice("somebody", "something")
	s.nc.Expect("NOTICE somebody :something")

	//                    01234567890123456789012345678901234567
	c.Notice("somebody", "something much much longer that splits")
	s.nc.Expect("NOTICE somebody :something much much ...")
	s.nc.Expect("NOTICE somebody :longer that splits")

	c.Ctcp("somebody", "ping", "123456789")
	s.nc.Expect("PRIVMSG somebody :\001PING 123456789\001")

	c.Ctcp("somebody", "ping", "123456789012345678901234567890")
	s.nc.Expect("PRIVMSG somebody :\001PING 12345678901234567890...\001")
	s.nc.Expect("PRIVMSG somebody :\001PING 1234567890\001")

	c.CtcpReply("somebody", "pong", "123456789012345678901234567890")
	s.nc.Expect("NOTICE somebody :\001PONG 12345678901234567890...\001")
	s.nc.Expect("NOTICE somebody :\001PONG 1234567890\001")

	c.CtcpReply("somebody", "pong", "123456789")
	s.nc.Expect("NOTICE somebody :\001PONG 123456789\001")

	c.Version("somebody")
	s.nc.Expect("PRIVMSG somebody :\001VERSION\001")

	c.Action("#foo", "pokes somebody")
	s.nc.Expect("PRIVMSG #foo :\001ACTION pokes somebody\001")

	c.Topic("#foo")
	s.nc.Expect("TOPIC #foo")
	c.Topic("#foo", "la la la")
	s.nc.Expect("TOPIC #foo :la la la")

	c.Mode("#foo")
	s.nc.Expect("MODE #foo")
	c.Mode("#foo", "+o somebody")
	s.nc.Expect("MODE #foo +o somebody")

	c.Away()
	s.nc.Expect("AWAY")
	c.Away("Dave's not here, man.")
	s.nc.Expect("AWAY :Dave's not here, man.")

	c.Invite("somebody", "#foo")
	s.nc.Expect("INVITE somebody #foo")

	c.Oper("user", "pass")
	s.nc.Expect("OPER user pass")

	c.VHost("user", "pass")
	s.nc.Expect("VHOST user pass")
}
