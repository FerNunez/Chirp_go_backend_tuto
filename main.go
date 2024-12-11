package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
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
	serverMux.HandleFunc("POST /api/validate_chirp", ValidateChirp)

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

func ValidateChirp(w http.ResponseWriter, r *http.Request) {

	type Chirp struct {
		Body string `json:"body"`
	}
	type ErrorResp struct {
		Error string `json:"error"`
	}
	type OkResp struct {
		CleanedBody string `json:"cleaned_body"`
	}

	// Decode received Json into Chirp
	chirp := Chirp{}
	err := json.NewDecoder(r.Body).Decode(&chirp)
	// If Error decoding chirp
	// Important to check that the body is empty cause could be an "error" when decoding json
	if err != nil || chirp.Body == "" {
		// Idk if I should ignore this err
		dat, _ := json.Marshal(ErrorResp{Error: "Something went wrong"})
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(500)
		w.Write([]byte(dat))
		return
	}

	// chrip is too long => respond with
	if len(chirp.Body) > 140 {
		errorResp := ErrorResp{Error: "Chrip is too long"}
		// Idk if I should ignore this err
		dat, _ := json.Marshal(errorResp)
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(500)
		w.Write([]byte(dat))
		return
	}

	// Ok chirp
	okResp := OkResp{CleanedBody: ReplaceBadWords(chirp.Body)}
	dat, _ := json.Marshal(okResp)
	// Idk if I should ignore the error from marshalin
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write([]byte(dat))
}

var badWords = []string{"kerfuffle", "sharbert", "fornax"}

func ReplaceBadWords(s string) string {
	string_array := strings.Split(s, " ")

	for index, w := range string_array {
		lowerCased := strings.ToLower(w)
		for _, badWord := range badWords {

			// Lower
			if lowerCased == badWord {
				string_array[index] = "****"
				break
			}
		}
	}
	return strings.Join(string_array, " ")

	//// Using replaceAll but it is not lower cased
	//sNew := strings.Clone(s)
	//for _, badWord := range []string{"kerfuffle", "sharbert", "fornax"} {
	//	sNew = strings.ReplaceAll(sNew, badWord, "****")
	//}
	//return sNew

}
