package client

import (
	"github.com/emersion/go-sasl"
	"testing"
)

func TestSaslPlainWorkflow(t *testing.T) {
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
