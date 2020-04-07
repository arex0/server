package server

import (
	"crypto/tls"
	"net/http"
	"reflect"
	"time"

	exbytes "github.com/arex0/go-ext/bytes"
)

// Server has mux and TLS config
type Server struct {
	mux       *Mux
	TLS       *TLS
	TLSConfig *tls.Config
	C         chan int
}

// Mux map mux.<handle,proxy>[port][path] -> <handle,proxy>
type Mux struct {
	handles        map[uint]map[string]func(w http.ResponseWriter, r *http.Request)
	defaultHandles map[uint]func(w http.ResponseWriter, r *http.Request)
}

// TLS contains two string pubKey,and priKey
type TLS struct {
	pubKey string
	priKey string
}

// Option export to user for easy extension
type Option func(*Server)

// WithTLS set https
func WithTLS(pubKey, priKey string) Option {
	return func(s *Server) {
		s.TLS = &TLS{
			pubKey: pubKey,
			priKey: priKey,
		}
	}
}

// WithTLSConfig set https by config
func WithTLSConfig(tls *tls.Config) Option {
	return func(s *Server) {
		s.TLSConfig = tls
	}
}

// New a Server
func New(mux *Mux, opts ...Option) *Server {
	//default options
	server := &Server{
		mux,
		nil,
		nil,
		make(chan int),
	}
	//set options
	for _, set := range opts {
		set(server)
	}
	return server
}

// Handle map mux.handles[port][path] -> handle
func (mux *Mux) Handle(port uint, s string, handle func(w http.ResponseWriter, r *http.Request)) {
	if mux.handles == nil {
		mux.handles = make(map[uint]map[string]func(w http.ResponseWriter, r *http.Request))
	}
	if _, ok := mux.handles[port]; !ok {
		mux.handles[port] = make(map[string]func(w http.ResponseWriter, r *http.Request))
	}
	if _, ok := mux.handles[port][s]; ok {
		panic("Mux.Handle failed:mux.handles[" + utoa(port) + "][" + s + "] has been rigister")
	}
	mux.handles[port][s] = handle
}

// DefaultHandle map mux.handles[port][path] -> handle
func (mux *Mux) DefaultHandle(port uint, handle func(w http.ResponseWriter, r *http.Request)) {
	if mux.defaultHandles == nil {
		mux.defaultHandles = make(map[uint]func(w http.ResponseWriter, r *http.Request))
	}
	if _, ok := mux.defaultHandles[port]; ok {
		panic("Mux.Handle failed:mux.handles[" + utoa(port) + "] has been rigister")
	}
	mux.defaultHandles[port] = handle
}

// Listen according to mux
func (s *Server) Listen(listen func()) {
	for _, port := range allKeys(s.mux.defaultHandles, s.mux.handles) {
		go func(port uint) {
			handle := httpHandle(s, port)
			if s.TLSConfig != nil {
				sev := &http.Server{
					Addr:         utoa(port),
					TLSConfig:    s.TLSConfig,
					Handler:      handle,
					ReadTimeout:  5 * time.Second,
					WriteTimeout: 10 * time.Second,
					IdleTimeout:  120 * time.Second,
				}
				sev.ListenAndServeTLS("", "")
			} else if s.TLS != nil {
				http.ListenAndServeTLS(":"+utoa(port), s.TLS.pubKey, s.TLS.priKey, handle)
			} else {
				http.ListenAndServe(":"+utoa(port), handle)
			}
		}(port)
	}
	listen()
}

func httpHandle(s *Server, port uint) http.HandlerFunc {
	handles := s.mux.handles[port]
	defaultHandle := s.mux.defaultHandles[port]
	return func(w http.ResponseWriter, r *http.Request) {
		selector := ParseSelector(r.URL.Path)
		if handle, ok := handles[selector]; ok {
			handle(w, r)
		} else {
			defaultHandle(w, r)
		}
	}
}

func allKeys(ms ...interface{}) []uint {
	fs := make(map[uint]bool)
	for _, m := range ms {
		for _, k := range reflect.ValueOf(m).MapKeys() {
			fs[uint(k.Uint())] = true
		}
	}
	var ks []uint
	for k := range fs {
		ks = append(ks, k)
	}
	return ks
}

func utoa(u uint) string {
	var b []byte
	if u == 0 {
		b = []byte{'0'}
	}
	for u > 0 {
		b = append(b, '0'+byte(u%10))
		u /= 10
	}
	return string(exbytes.Reverse(b))
}
