package client

import (
	"net"
	"os"
	"strings"
	"testing"
	"time"
)

type mockNetConn struct {
	*testing.T

	In, Out chan string
	in, out chan []byte
	closers []chan bool
	rc chan bool

	closed bool
	rt, wt int64
}

func MockNetConn(t *testing.T) (*mockNetConn) {
	// Our mock connection is a testing object
	m := &mockNetConn{T: t}
	m.closers = make([]chan bool, 0, 3)

	// set known values for conn info
	m.closed = false
	m.rt = 0
	m.wt = 0

	// buffer input
	m.In = make(chan string, 20)
	m.in = make(chan []byte)
	ic := make(chan bool)
	m.closers = append(m.closers, ic)
	go func() {
		for {
			select {
			case <-ic:
				return
			case s := <-m.In:
				m.in <- []byte(s)
			}
		}
	}()

	// buffer output
	m.Out = make(chan string)
	m.out = make(chan []byte, 20)
	oc := make(chan bool)
	m.closers = append(m.closers, oc)
	go func() {
		for {
			select {
			case <-oc:
				return
			case b := <-m.out:
				m.Out <- string(b)
			}
		}
	}()

	// Set up channel to force EOF to Read() on close.
	m.rc = make(chan bool)
	m.closers = append(m.closers, m.rc)

	return m
}

// Test helpers
func (m *mockNetConn) Send(s string) {
	m.In <- s + "\r\n"
}

func (m *mockNetConn) Expect(e string) {
	t := time.NewTimer(5e6)
	select {
	case <-t.C:
		m.Errorf("Mock connection did not receive expected output.\n\t" +
			"Expected: '%s', got nothing.", e)
	case s := <-m.Out:
		t.Stop()
		s = strings.Trim(s, "\r\n")
		if e != s {
			m.Errorf("Mock connection received unexpected value.\n\t" +
				"Expected: '%s'\n\tGot: '%s'", e, s)
		}
	}
}

func (m *mockNetConn) ExpectNothing() {
	t := time.NewTimer(5e6)
	select {
	case <-t.C:
	case s := <-m.Out:
		t.Stop()
		s = strings.Trim(s, "\r\n")
		m.Errorf("Mock connection received unexpected output.\n\t" +
			"Expected nothing, got: '%s'", s)
	}
}

// Implement net.Conn interface
func (m *mockNetConn) Read(b []byte) (int, os.Error) {
	if m.closed {
		return 0, os.EINVAL
	}
	select {
	case s := <-m.in:
		copy(b, s)
	case <-m.rc:
		return 0, os.EOF
	}
	return len(b), nil
}

func (m *mockNetConn) Write(s []byte) (int, os.Error) {
	if m.closed {
		return 0, os.EINVAL
	}
	b := make([]byte, len(s))
	copy(b, s)
	m.out <- b
	return len(b), nil
}

func (m *mockNetConn) Close() os.Error {
	if m.closed {
		return os.EINVAL
	}
	// Shut down *ALL* the goroutines!
	// This will trigger an EOF event in Read() too
	for _, c := range m.closers {
		c <- true
	}
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
