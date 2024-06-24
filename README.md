# go-httprawdump

Utilities to print reads and writes to `net.Conn`.
Useful for debugging when the environment makes it impractical to get tcpdump going.


## Wiretap Listener

A listener that wraps calls to `Read` and `Write` and tees the bytes to an `io.Writer`.

Example usage:

```go
package main
import (
    "log"
    "net"
    "net/http"
    "os"

    "github.com/brittlesoft/go-httprawdump/pkg/wiretap"
)

func main() {
    l, err := net.Listen("tcp", "[::1]:0")
    if err != nil {
        panic(err)
    }
    wiretapListener := wiretap.Listener{Wrapped: l, OutputWriter: os.Stdout}

    log.Printf("Listening on: %s", wiretapListener.Addr())
    handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        _, err := w.Write([]byte("ok"))
        if err != nil {
            panic(err)
        }
    })

    http.Serve(wiretapListener, handler)
}
```

Output:
```
2024/06/23 22:58:04 Listening on: [::1]:34079
[::1]:34079 [::1]:52632 read begin
GET / HTTP/1.1
Host: localhost:34079
User-Agent: curl/8.8.0
Accept: */*


[::1]:34079 [::1]:52632 read end
[::1]:34079 [::1]:52632 write begin
HTTP/1.1 200 OK
Date: Mon, 24 Jun 2024 02:58:11 GMT
Content-Length: 2
Content-Type: text/plain; charset=utf-8

ok
[::1]:34079 [::1]:52632 write end
```

Nothing fancy.

See the tests and  `examples/wiretap` directory for more.
