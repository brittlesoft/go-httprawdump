package wiretap

import (
	"fmt"
	"io"
	"net"
)

type conn struct {
	net.Conn
	out        io.Writer
	outReadCh  chan []byte
	outWriteCh chan []byte
	doneCh     chan struct{}
}

func newConn(inconn net.Conn, out io.Writer) *conn {
	c := &conn{
		Conn:       inconn,
		out:        out,
		outReadCh:  make(chan []byte),
		outWriteCh: make(chan []byte),
		doneCh:     make(chan struct{}, 2), // allow for both sides to close (testing)
	}
	c.startOut()

	return c
}

func (c *conn) startOut() {
	rIdent := []byte(fmt.Sprintf("%s read\n", c.RemoteAddr()))
	wIdent := []byte(fmt.Sprintf("%s write\n", c.RemoteAddr()))
	go func() {
		for {
			select {
			case b := <-c.outReadCh:
				c.out.Write(rIdent)
				c.out.Write(b)
			case b := <-c.outWriteCh:
				c.out.Write(wIdent)
				c.out.Write(b)
			case <-c.doneCh:
				return
			}
		}
	}()
}

func (c *conn) Read(b []byte) (n int, err error) {
	n, err = c.Conn.Read(b)
	if err != nil {
		return 0, err
	}

	cb := make([]byte, len(b))
	copy(cb, b)
	c.outReadCh <- cb
	return n, err
}

func (c *conn) Write(b []byte) (n int, err error) {
	n, err = c.Conn.Write(b)
	if err != nil {
		return 0, err
	}

	cb := make([]byte, len(b))
	copy(cb, b)
	c.outWriteCh <- cb
	return n, err
}

func (c *conn) Close() error {
	c.doneCh <- struct{}{}
	return c.Conn.Close()
}

type Listener struct {
	Wrapped      net.Listener
	OutputWriter io.Writer
}

func (l Listener) Accept() (net.Conn, error) {
	conn, err := l.Wrapped.Accept()
	if err != nil {
		return nil, err
	}
	wtConn := newConn(conn, l.OutputWriter)
	return wtConn, err
}
func (l Listener) Close() error {
	return l.Wrapped.Close()
}
func (l Listener) Addr() net.Addr {
	return l.Wrapped.Addr()
}
