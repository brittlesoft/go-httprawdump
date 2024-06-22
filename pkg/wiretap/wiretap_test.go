package wiretap

import (
	"log"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestWiretap(t *testing.T) {
	s := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte{})
	}))

	l, err := net.Listen("tcp", "[::1]:0")
	if err != nil {
		t.Error(err)
	}
	s.Listener = Listener{Wrapped: l, OutputWriter: os.Stdout}
	s.Start()
	defer s.Close()

	resp, err := http.Get(s.URL)
	if err != nil {
		t.Errorf("Get failed: %v", err)
	}

	log.Printf("resp: %+v", resp)

}
