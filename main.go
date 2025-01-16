// The Go standard library makes it easy to build a simple server. Your task is to build and run a server that binds to localhost:8080 and always responds with a 404 Not Found response.
package main

import (
	"net/http"
	"strconv"
	"sync/atomic"
)

type apiConfig struct {
	fileserverHits atomic.Int32
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {

	// CHECK AND LEARN THIS??
	// create wrapping handler Func
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Increment 1
		cfg.fileserverHits.Add(1)
		// call wrapped handler
		next.ServeHTTP(w, r)
	})

}

func main() {

	cfg := apiConfig{}
	serverMux := http.NewServeMux()
	serverMux.Handle("/", http.StripPrefix("/app/", cfg.middlewareMetricsInc(http.FileServer(http.Dir(".")))))
	serverMux.Handle("assets/logo.png", http.FileServer(http.Dir("assets/logo.png")))
	serverMux.HandleFunc("GET /healthz", readinnesHandler)
	serverMux.HandleFunc("GET /metrics", cfg.metricsDisplayHandler)
	serverMux.HandleFunc("POST /reset", cfg.metricsResetHandler)

	// Listen & Serve
	server := &http.Server{
		Handler: serverMux,
		Addr:    ":8080",
	}
	server.ListenAndServe()

}

func readinnesHandler(rw http.ResponseWriter, r *http.Request) {
	rw.Header().Add("Content-Type", "text/plain;charset=utf-8")
	rw.WriteHeader(200)
	rw.Write([]byte("Ok"))
}

func (cfg *apiConfig) metricsDisplayHandler(rw http.ResponseWriter, _ *http.Request) {
	rw.Header().Add("Content-Type", "text/plain;charset=utf-8")
	rw.WriteHeader(200)

	x := int(cfg.fileserverHits.Load())
	rw.Write([]byte(strconv.Itoa(x)))

}

func (cfg *apiConfig) metricsResetHandler(rw http.ResponseWriter, _ *http.Request) {
	rw.Header().Add("Content-Type", "text/plain;charset=utf-8")
	rw.WriteHeader(200)

	cfg.fileserverHits.Store(0)
	rw.Write([]byte("Counter Reseted"))

}
