package main

import (
	"fmt"
	"net/http"
)

func main() {

	serverMux := http.NewServeMux()

	// Create server
	server := http.Server{
		Handler: serverMux,
		Addr:    ":8080",
	}

	// add handle to / path by looking current directory
	serverMux.Handle("/", http.FileServer(http.Dir(".")))

	// launch server
	err := server.ListenAndServe()
	if err != nil {
		fmt.Printf("Error encountered launching server %v", err)
		return
	}

}
