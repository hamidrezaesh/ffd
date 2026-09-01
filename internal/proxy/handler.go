package proxy

import (
	"io"
	"log"
	"net"
	"net/http"
	"time"
)

func (s *Server) handleRequest(w http.ResponseWriter, r *http.Request) {
	outReq := r.Clone(r.Context())
	outReq.RequestURI = ""

	if outReq.URL.Scheme == "" {
		outReq.URL.Scheme = "http"
	}

	if outReq.URL.Host == "" {
		outReq.URL.Host = r.Host
	}

	// Ask the upstream server for metadata.
	headReq := outReq.Clone(r.Context())
	headReq.Method = http.MethodHead
	headReq.Body = nil
	headReq.ContentLength = 0

	headResp, err := http.DefaultClient.Do(headReq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	defer headResp.Body.Close()

	totalSize := headResp.ContentLength
	acceptRanges := headResp.Header.Get("Accept-Ranges") == "bytes"

	// use the normal proxy path for small files or servers that doesnt support accept range
	if totalSize <= 0 ||
		totalSize < accelerationThreshold ||
		!acceptRanges {

		s.proxyNormal(w, outReq)
		return
	}

	// Large range-supported resource.
	s.proxyAccelerated(
		w,
		outReq,
		headResp.Header,
		totalSize,
	)
}

func (s *Server) handleConnect(w http.ResponseWriter, r *http.Request) {
	destConn, err := net.DialTimeout("tcp", r.Host, 10*time.Second)
	if err != nil {
		http.Error(w, "Failed to connect to destination", http.StatusServiceUnavailable)
		return
	}
	defer destConn.Close()

	hijacker, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "Hijacking not supported", http.StatusInternalServerError)
		return
	}

	clientConn, _, err := hijacker.Hijack()
	if err != nil {
		log.Printf("Hijack failed: %v", err)
		return
	}
	defer clientConn.Close()

	_, err = clientConn.Write([]byte(
		"HTTP/1.1 200 Connection Established\r\n\r\n",
	))
	if err != nil {
		return
	}

	go io.Copy(destConn, clientConn)
	io.Copy(clientConn, destConn)

	log.Printf("CONNECT tunnel closed for %s", r.Host)
}
