package client

import (
	"github.com/emersion/go-sasl"
	"testing"
)

func TestSaslPlainSuccessWorkflow(t *testing.T) {
	c, s := setUp(t)
	defer s.tearDown()

	c.Config().Sasl = sasl.NewPlainClient("", "example", "password")
	c.Config().EnableCapabilityNegotiation = true

	c.h_REGISTER(&Line{Cmd: REGISTER})
	s.nc.Expect("CAP LS")
	s.nc.Expect("NICK test")
	s.nc.Expect("USER test 12 * :Testing IRC")
	s.nc.Send("CAP * LS :sasl foobar")
	s.nc.Expect("CAP REQ :sasl")
	s.nc.Send("CAP * ACK :sasl")
	s.nc.Expect("AUTHENTICATE PLAIN")
	s.nc.Send("AUTHENTICATE +")
	s.nc.Expect("AUTHENTICATE AGV4YW1wbGUAcGFzc3dvcmQ=")
	s.nc.Send("904 test :SASL authentication successful")
	s.nc.Expect("CAP END")
}

func TestSaslPlainWrongPassword(t *testing.T) {
	c, s := setUp(t)
	defer s.tearDown()

	c.Config().Sasl = sasl.NewPlainClient("", "example", "password")
	c.Config().EnableCapabilityNegotiation = true

	c.h_REGISTER(&Line{Cmd: REGISTER})
	s.nc.Expect("CAP LS")
	s.nc.Expect("NICK test")
	s.nc.Expect("USER test 12 * :Testing IRC")
	s.nc.Send("CAP * LS :sasl foobar")
	s.nc.Expect("CAP REQ :sasl")
	s.nc.Send("CAP * ACK :sasl")
	s.nc.Expect("AUTHENTICATE PLAIN")
	s.nc.Send("AUTHENTICATE +")
	s.nc.Expect("AUTHENTICATE AGV4YW1wbGUAcGFzc3dvcmQ=")
	s.nc.Send("904 test :SASL authentication failed")
	s.nc.Expect("CAP END")
}

func TestSaslExternalSuccessWorkflow(t *testing.T) {
	c, s := setUp(t)
	defer s.tearDown()

	c.Config().Sasl = sasl.NewExternalClient("")
	c.Config().EnableCapabilityNegotiation = true

	c.h_REGISTER(&Line{Cmd: REGISTER})
	s.nc.Expect("CAP LS")
	s.nc.Expect("NICK test")
	s.nc.Expect("USER test 12 * :Testing IRC")
	s.nc.Send("CAP * LS :sasl foobar")
	s.nc.Expect("CAP REQ :sasl")
	s.nc.Send("CAP * ACK :sasl")
	s.nc.Expect("AUTHENTICATE EXTERNAL")
	s.nc.Send("AUTHENTICATE +")
	s.nc.Expect("AUTHENTICATE +")
	s.nc.Send("904 test :SASL authentication successful")
	s.nc.Expect("CAP END")
}

func TestSaslNoSaslCap(t *testing.T) {
	c, s := setUp(t)
	defer s.tearDown()

	c.Config().Sasl = sasl.NewPlainClient("", "example", "password")
	c.Config().EnableCapabilityNegotiation = true

	c.h_REGISTER(&Line{Cmd: REGISTER})
	s.nc.Expect("CAP LS")
	s.nc.Expect("NICK test")
	s.nc.Expect("USER test 12 * :Testing IRC")
	s.nc.Send("CAP * LS :foobar")
	s.nc.Expect("CAP END")
}

func TestSaslUnsupportedMechanism(t *testing.T) {
	c, s := setUp(t)
	defer s.tearDown()

	c.Config().Sasl = sasl.NewPlainClient("", "example", "password")
	c.Config().EnableCapabilityNegotiation = true

	c.h_REGISTER(&Line{Cmd: REGISTER})
	s.nc.Expect("CAP LS")
	s.nc.Expect("NICK test")
	s.nc.Expect("USER test 12 * :Testing IRC")
	s.nc.Send("CAP * LS :sasl foobar")
	s.nc.Expect("CAP REQ :sasl")
	s.nc.Send("CAP * ACK :sasl")
	s.nc.Expect("AUTHENTICATE PLAIN")
	s.nc.Send("908 test external :are available SASL mechanisms")
	s.nc.Expect("CAP END")
}
