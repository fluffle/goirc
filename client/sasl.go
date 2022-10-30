package client

import (
	"encoding/base64"
)

// saslMechanism is the name of the SASL authentication mechanism used.
type saslMechanism string

const (
	// saslPlain is the username and password based PLAIN
	// authentication mechanism.
	saslPlain saslMechanism = "PLAIN"
)

// saslCap is the IRCv3 capability used for SASL authentication.
const saslCap = "sasl"

// SaslAuthenticator authenticates the connection using SASL in the
// connection phase.
type SaslAuthenticator struct {
	mechanism             saslMechanism
	authenticationRequest func() string
}

func encodePlainUsernamePassword(username, password string) string {
	requestBytes := []byte(username)
	requestBytes = append(requestBytes, byte(0))
	requestBytes = append(requestBytes, []byte(username)...)
	requestBytes = append(requestBytes, byte(0))
	requestBytes = append(requestBytes, []byte(password)...)

	return base64.StdEncoding.EncodeToString(requestBytes)
}

func SaslPlain(username, password string) *SaslAuthenticator {
	return &SaslAuthenticator{
		mechanism: saslPlain,
		authenticationRequest: func() string {
			return encodePlainUsernamePassword(username, password)
		},
	}
}
