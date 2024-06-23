package wiretap

import (
	"bytes"
	"crypto/rand"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"testing"
)

func TestWiretap(t *testing.T) {
	s := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Error(err)
		}

		_, err = w.Write(body)
		if err != nil {
			t.Error(err)
		}
	}))

	l, err := net.Listen("tcp", "[::1]:0")
	if err != nil {
		t.Error(err)
	}
	s.Listener = Listener{Wrapped: l, OutputWriter: os.Stdout}
	s.Start()
	defer s.Close()

	wantBody := "test string 1 2 3"
	resp, err := http.Post(s.URL, "something/orother", bytes.NewBufferString(wantBody))
	if err != nil {
		t.Errorf("req failed: %v", err)
	}
	gotBody, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Error(err)
	}
	if string(gotBody) != wantBody {
		t.Errorf("want: %s, got: %s", wantBody, gotBody)
	}
}

func echo(l net.Listener, payload []byte, b *testing.B) {
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		conn, err := l.Accept()
		if err != nil {
			panic(err)
		}
		defer conn.Close()

		io.Copy(conn, conn)
		wg.Done()
	}()

	conn, err := net.Dial(l.Addr().Network(), l.Addr().String())
	if err != nil {
		b.Error(err)
	}

	readb := make([]byte, len(payload))
	for range b.N {
		_, err := conn.Write(payload)
		if err != nil {
			b.Error(err)
		}
		_, err = conn.Read(readb)
		if err != nil {
			b.Error(err)
		}
	}
	conn.Close()
	wg.Wait()
}
func tcplistener() net.Listener {
	l, err := net.Listen("tcp", "[::1]:0")
	if err != nil {
		panic(err)
	}
	return l
}
func BenchmarkTCPPlain8b(b *testing.B) {
	echo(tcplistener(), []byte("patateee"), b)
}
func BenchmarkTCPWiretapDiscard8b(b *testing.B) {
	echo(Listener{tcplistener(), io.Discard}, []byte("patate"), b)
}
func BenchmarkTCPPlain8k(b *testing.B) {
	var payload [8192]byte
	_, err := rand.Read(payload[:])
	if err != nil {
		panic(err)
	}
	echo(tcplistener(), payload[:], b)
}
func BenchmarkTCPWiretapDiscard8k(b *testing.B) {
	var payload [8192]byte
	_, err := rand.Read(payload[:])
	if err != nil {
		panic(err)
	}
	echo(Listener{tcplistener(), io.Discard}, payload[:], b)
}
