package proxy

import (
	"fmt"
	"log"
	"net/http"
)

type Server struct {
	Addr   string
	Client *http.Client
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodConnect {
		s.handleConnect(w, r)
		return
	}

	s.handleRequest(w, r)
}

func Start(port int) {
	// set port to default port if not exists
	defaultPort := 8000
	if port == 0 {
		port = defaultPort
	}

	server := &Server{
		Addr:   fmt.Sprintf("127.0.0.1:%v", port),
		Client: &http.Client{},
	}

	httpServer := &http.Server{
		Addr:    server.Addr,
		Handler: server,
	}

	log.Println("ffd proxy listening on ", server.Addr)

	if err := httpServer.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
