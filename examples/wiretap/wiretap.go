package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"

	"github.com/brittlesoft/go-httprawdump/pkg/wiretap"
)

func main() {
	s := http.Server{
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, err := w.Write([]byte("ok"))
			if err != nil {
				panic(err)
			}
		}),
	}

	l, err := net.Listen("tcp", "[::1]:0")
	if err != nil {
		panic(err)
	}

	wiretapListener := wiretap.Listener{Wrapped: l, OutputWriter: os.Stdout}

	u := fmt.Sprintf("http://%s", wiretapListener.Addr())
	log.Printf("Listening on: %s", wiretapListener.Addr())
	go s.Serve(wiretapListener)

	http.Get(u)
}
