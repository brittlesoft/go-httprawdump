package wiretap

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"sync"
)

var bufPool = sync.Pool{
	New: func() any {
		return new(bytes.Buffer)
	},
}

type conn struct {
	net.Conn
	out        io.Writer
	outReadCh  chan *bytes.Buffer
	outWriteCh chan *bytes.Buffer
	doneCh     chan struct{}
	pool       *sync.Pool
}

func newConn(inconn net.Conn, out io.Writer) *conn {
	c := &conn{
		Conn:       inconn,
		out:        out,
		outReadCh:  make(chan *bytes.Buffer),
		outWriteCh: make(chan *bytes.Buffer),
		doneCh:     make(chan struct{}, 2), // allow for both sides to close (testing)
		pool:       &bufPool,
	}
	c.startOut()

	return c
}

func (c *conn) startOut() {
	rIdent := fmt.Sprintf("%s %s read", c.LocalAddr(), c.RemoteAddr())
	wIdent := fmt.Sprintf("%s %s write", c.LocalAddr(), c.RemoteAddr())
	outBuf := bytes.Buffer{}

	writeOut := func(rcvBuf *bytes.Buffer, ident *string) {
		outBuf.Reset()
		outBuf.WriteString(*ident)
		outBuf.WriteString(" begin\n")

		outBuf.Write(rcvBuf.Bytes())

		outBuf.WriteString("\n")
		outBuf.WriteString(*ident)
		outBuf.WriteString(" end\n")
		c.out.Write(outBuf.Bytes()) // TODO: log on error?
		c.pool.Put(rcvBuf)
	}

	go func() {
		for {
			select {
			case b := <-c.outReadCh:
				writeOut(b, &rIdent)
			case b := <-c.outWriteCh:
				writeOut(b, &wIdent)
			case <-c.doneCh:
				return
			}
		}
	}()
}

func (c *conn) sendToWriter(b []byte, ch chan *bytes.Buffer) {
	buf := c.pool.Get().(*bytes.Buffer)
	buf.Reset()
	buf.Write(b)
	ch <- buf
}

func (c *conn) Read(b []byte) (int, error) {
	n, err := c.Conn.Read(b)
	if err == nil {
		c.sendToWriter(b[:n], c.outReadCh)
	}
	return n, err
}

func (c *conn) Write(b []byte) (int, error) {
	n, err := c.Conn.Write(b)
	if err == nil {
		c.sendToWriter(b[:n], c.outWriteCh)
	}

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
