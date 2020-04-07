a simple golang server
```go
package main

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/arex0/server"
	"golang.org/x/crypto/acme/autocert"
)

func main() {
	var mux server.Mux
	counter := make(map[string]int)
	mux.Handle(443, "visited_count", func(w http.ResponseWriter, r *http.Request) {
		if len(r.URL.Query()) != 0 {
			counter[r.URL.Path]++
		}
		w.Header().Set("Content-Type", "image/svg+xml; charset=utf-8")
		w.Write([]byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 4 1"><text x="0" y="1em" font-size="1" font-family="monaco,consolas,monospace">` + strconv.Itoa(counter[r.URL.Path]) + `</text></svg>`))
	})
	mux.Handle(443, "like_count", func(w http.ResponseWriter, r *http.Request) {
		if len(r.URL.Query()) != 0 {
			counter[r.URL.Path]++
		}
		w.Header().Set("Content-Type", "image/svg+xml; charset=utf-8")
		w.Write([]byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 4 1"><text x="0" y="1em" font-size="1" font-family="monaco,consolas,monospace">` + strconv.Itoa(counter[r.URL.Path]) + `</text></svg>`))
	})
	mux.Handle(443, "visit_like_count", func(w http.ResponseWriter, r *http.Request) {
		var b bytes.Buffer
		for k, v := range counter {
			b.WriteString("{\"" + k + "\":" + strconv.Itoa(v) + "}\n")
		}
		w.Write(b.Bytes())
	})
	certManager := autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist("arex0.com"),
		Cache:      autocert.DirCache("cert"),
		Email:      "cn.js.cross@gmail.com",
	}
	server := server.New(&mux, server.WithTLSConfig(tls.Config{
		GetCertificate:           certManager.GetCertificate,
		NextProtos:               []string{"h2", "http/1.1"},
		MinVersion:               tls.VersionTLS12,
		PreferServerCipherSuites: false,
		CipherSuites:             []uint16{},
	}))
	go func() {
		var C int
		for {
			fmt.Scanln(&C)
			server.C <- C
		}
	}()
	go http.ListenAndServe(":80", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Strict-Transport-Security", "max-age=63072000")
		w.Header().Set("Connection", "close")
		http.Redirect(w, r, "https://"+r.Host+r.URL.String(), 301)
	}))
	server.Listen(func() {
		for {
			switch <-server.C {
			case 0:
				os.Exit(0)
			}
		}
	})
}
```
