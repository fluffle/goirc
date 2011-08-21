package client

import (
	"net"
	"os"
	"testing"
)

type mockNetConn struct {
	*testing.T

	In, Out chan string
	in, out chan []byte

	closed bool
	rt, wt int64
}

func MockNetConn(t *testing.T) (*mockNetConn) {
	// Our mock connection is a testing object
	m := &mockNetConn{T: t}

	// set known values for conn info
	m.closed = false
	m.rt = 0
	m.wt = 0

	// buffer input
	m.In = make(chan string, 20)
	m.in = make(chan []byte)
	go func() {
		for !m.closed {
			m.in <- []byte(<-m.In)
		}
	}()

	// buffer output
	m.Out = make(chan string)
	m.out = make(chan []byte, 20)
	go func() {
		for !m.closed {
			m.Out <- string(<-m.out)
		}
	}()
	return m
}

// Test helper
func (m *mockNetConn) Expect(e string) {
	s := <-m.Out
	if e + "\r\n" != s {
		m.Errorf("Mock connection received unexpected value.\n\t" +
			"Expected: %s\n\tGot: %s", e, s)
	}
}

// Implement net.Conn interface
func (m *mockNetConn) Read(b []byte) (int, os.Error) {
	s := <-m.in
	copy(b, s)
	return len(s), nil
}

func (m *mockNetConn) Write(s []byte) (int, os.Error) {
	b := make([]byte, len(s))
	copy(b, s)
	m.out <- b
	return len(b), nil
}

func (m *mockNetConn) Close() os.Error {
	m.closed = true
	return nil
}

func (m *mockNetConn) LocalAddr() net.Addr {
	return &net.IPAddr{net.IPv4(127,0,0,1)}
}

func (m *mockNetConn) RemoteAddr() net.Addr {
	return &net.IPAddr{net.IPv4(127,0,0,1)}
}

func (m *mockNetConn) SetTimeout(ns int64) os.Error {
	m.rt = ns
	m.wt = ns
	return nil
}

func (m *mockNetConn) SetReadTimeout(ns int64) os.Error {
	m.rt = ns
	return nil
}

func (m *mockNetConn) SetWriteTimeout(ns int64) os.Error {
	m.wt = ns
	return nil
}
