package main

import (
	"fmt"
	"net/http"
)

func main() {

	serverMux := http.NewServeMux()

	server := http.Server{
		Handler: serverMux,
		Addr:    ":8080",
	}
	err := server.ListenAndServe()
	if err != nil {

		fmt.Printf("Error encountered launching server %v", err)
		return
	}

}
