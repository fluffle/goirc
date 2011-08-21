package client

import "testing"

func TestClientCommands(t *testing.T) {
	m, c := setUp(t)

	c.Pass("password")
	m.Expect("PASS password")

	c.Nick("test")
	m.Expect("NICK test")

	c.User("test", "Testing IRC")
	m.Expect("USER test 12 * :Testing IRC")

	c.Raw("JUST a raw :line")
	m.Expect("JUST a raw :line")

	c.Join("#foo")
	m.Expect("JOIN #foo")

	c.Part("#foo")
	m.Expect("PART #foo")
	c.Part("#foo", "Screw you guys...")
	m.Expect("PART #foo :Screw you guys...")

	c.Quit()
	m.Expect("QUIT :GoBye!")
	c.Quit("I'm going home.")
	m.Expect("QUIT :I'm going home.")

	c.Whois("somebody")
	m.Expect("WHOIS somebody")

	c.Who("*@some.host.com")
	m.Expect("WHO *@some.host.com")

	c.Privmsg("#foo", "bar")
	m.Expect("PRIVMSG #foo :bar")

	c.Notice("somebody", "something")
	m.Expect("NOTICE somebody :something")

	c.Ctcp("somebody", "ping", "123456789")
	m.Expect("PRIVMSG somebody :\001PING 123456789\001")

	c.CtcpReply("somebody", "pong", "123456789")
	m.Expect("NOTICE somebody :\001PONG 123456789\001")

	c.Version("somebody")
	m.Expect("PRIVMSG somebody :\001VERSION\001")

	c.Action("#foo", "pokes somebody")
	m.Expect("PRIVMSG #foo :\001ACTION pokes somebody\001")

	c.Topic("#foo")
	m.Expect("TOPIC #foo")
	c.Topic("#foo", "la la la")
	m.Expect("TOPIC #foo :la la la")

	c.Mode("#foo")
	m.Expect("MODE #foo")
	c.Mode("#foo", "+o somebody")
	m.Expect("MODE #foo +o somebody")

	c.Away()
	m.Expect("AWAY")
	c.Away("Dave's not here, man.")
	m.Expect("AWAY :Dave's not here, man.")

	c.Invite("somebody", "#foo")
	m.Expect("INVITE somebody #foo")

	c.Oper("user", "pass")
	m.Expect("OPER user pass")
}
