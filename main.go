package main

import (
	"fmt"
	"net/http"
	"sync/atomic"
)

func main() {

	serverMux := http.NewServeMux()

	// Create server
	server := http.Server{
		Handler: serverMux,
		Addr:    ":8080",
	}
	apiCfg := apiConfig{}
	// add handles
	serverMux.Handle("/app/", apiCfg.middlewareMetricsInc(http.StripPrefix("/app/", http.FileServer(http.Dir(".")))))
	serverMux.HandleFunc("GET /api/healthz", ReadinessHandler)
	serverMux.HandleFunc("GET /admin/metrics", apiCfg.NumberOfRequests)
	serverMux.HandleFunc("POST /admin/reset", apiCfg.ResetNumberOfRequests)

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

type apiConfig struct {
	fileserverHits atomic.Int32
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {

	// create wrapping handler Func
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Increment 1
		cfg.fileserverHits.Add(1)
		// call wrapped handler
		next.ServeHTTP(w, r)

	})
}

func (cfg *apiConfig) NumberOfRequests(w http.ResponseWriter, r *http.Request) {
	httpResponse := fmt.Sprintf(`
	<html>
		<body>
			<h1>Welcome, Chirpy Admin</h1>
			<p>Chirpy has been visited %d times!</p>
		</body>
	</html>
`, cfg.fileserverHits.Load())

	//w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.Header().Add("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK) // I did 200 instead of http.StatusOk
	fmt.Fprintf(w, "%v", httpResponse)
}
func (cfg *apiConfig) ResetNumberOfRequests(w http.ResponseWriter, r *http.Request) {
	cfg.fileserverHits.Store(0)

	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK) // I did 200 instead of http.StatusOk
	fmt.Fprintf(w, "Reset done: %v", cfg.fileserverHits.Load())
}
