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

	// add handles
	serverMux.Handle("/app/", http.StripPrefix("/app/", http.FileServer(http.Dir("."))))
	serverMux.HandleFunc("/healthz", ReadinessHandler)

	// launch server
	err := server.ListenAndServe()
	if err != nil {
		fmt.Printf("Error encountered launching server %v", err)
		return
	}

}

// Readiness: sends a OK
func ReadinessHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK) // I did 200 instead of http.StatusOk
	w.Write([]byte("OK"))
}
