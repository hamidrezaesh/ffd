package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
)

/*
Test Server

This is a local HTTP server used for testing and debugging FFD.

It prints information about every incoming request, including:
- HTTP method
- URL
- HTTP protocol
- Remote address
- Host
- Request headers
- Request body

It can be used to test features such as:
- Custom HTTP headers
- HTTP protocol selection
- Range requests
- Proxy behavior
- Other HTTP client functionality

Run from the project root with:

	go run test/test-server.go

The server listens on:

	http://localhost:8080

then you can run tests such as
	ffd --header="User-Agent: FFD" http://localhost:8080
to see if the ffd's new feature is working or not

This is a development and debugging tool and is not intended for production use.
*/

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("========================================")
	fmt.Println("Request received")
	fmt.Println("========================================")

	fmt.Println("Method:        ", r.Method)
	fmt.Println("URL:           ", r.URL.String())
	fmt.Println("Protocol:      ", r.Proto)
	fmt.Println("Remote address:", r.RemoteAddr)
	fmt.Println("Host:          ", r.Host)

	fmt.Println("\nHeaders:")
	for key, values := range r.Header {
		for _, value := range values {
			fmt.Printf("  %s: %s\n", key, value)
		}
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		fmt.Println("\nBody error:", err)
	} else {
		fmt.Println("\nBody size:     ", len(body), "bytes")

		if len(body) > 0 {
			fmt.Println("Body:")
			fmt.Println(string(body))
		}
	}

	fmt.Println("========================================")
	fmt.Println()

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("Hello from test server!\n"))
}

func main() {
	http.HandleFunc("/", handler)

	addr := ":8080"

	log.Printf("Test server listening on http://localhost%s", addr)

	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal(err)
	}
}
