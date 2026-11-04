package proxy

import (
	"log"
	"net/http"
	"net/url"

	"github.com/hamidrezaesh/ffd/internal/scheduler"
)

type Server struct {
	Addr              string
	Proxy             *url.URL
	Headers           scheduler.Headers
	CookieJar         http.CookieJar
	MaxRetries        int
	MaxWorkers        int
	MaxChunks         int
	PreferredProtocol int
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodConnect {
		s.handleConnect(w, r)
		return
	}

	s.handleRequest(w, r)
}

func Start(s Server) {
	// Set default address if not provided.
	if s.Addr == "" {
		s.Addr = "127.0.0.1:8000"
	}

	// Set defaults.
	if s.MaxRetries <= 0 {
		s.MaxRetries = 4
	}
	if s.MaxWorkers <= 0 {
		s.MaxWorkers = 8
	}
	if s.MaxChunks <= 0 {
		s.MaxChunks = 12
	}

	httpServer := &http.Server{
		Addr:    s.Addr,
		Handler: &s,
	}

	log.Println("FFD proxy listening on", s.Addr)

	if err := httpServer.ListenAndServe(); err != nil &&
		err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
