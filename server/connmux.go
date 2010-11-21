package server

import (
	"bufio"
	"net"
	"crypto/tls"
	"compress/zlib"
)

type Mux struct {
	// Sockets we've created
	listening []net.Listener
	serving   []net.Conn

	// send and recieve channels to sockets
	send map[*Node]chan string
	recv chan string

	// input/output/error channels to daemon
	// Msg interface defined in parser.go
	In  chan *Msg
	Out chan *Msg
	Err chan os.Error
}

func (m *Mux) Serve(addr string, client bool, conf *tls.Config) os.Error {
	var l net.Listener, s net.Conn, e os.Error
	if conf == nil {
		if l, e = net.Listen("tcp", addr); e != nil {
			return e
		}
	} else {
		if l, e = tls.Listen("tcp", addr, conf); e != nil {
			return e
		}
	}
	append(m.listening, l)
	go func() {
		for {
			if s, e = l.Accept(); e != nil {
				m.Err <- e
			}
			append(m.serving, s)
			io := bufio.NewReadWriter(bufio.NewReader(s), bufio.NewWriter(s))
			if client {
				go m.clientSync(io)
			} else {
				// TODO(abramley): zlib support
				go m.serverSync(io)
			}
		}
	}
}

func (m *Mux) clientSync(in bufio.ReadWriter) {
	
}

