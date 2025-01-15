package main

import (
	"net/http"
)

//The Go standard library makes it easy to build a simple server. Your task is to build and run a server that binds to localhost:8080 and always responds with a 404 Not Found response.

func main() {

	serverMux := http.NewServeMux()
	serverMux.Handle("/", http.FileServer(http.Dir(".")))

	// Listen & Serve
	server := &http.Server{
		Handler: serverMux,
		Addr:    ":8080",
	}
	server.ListenAndServe()

}
