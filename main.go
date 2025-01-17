// The Go standard library makes it easy to build a simple server. Your task is to build and run a server that binds to localhost:8080 and always responds with a 404 Not Found response.
package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync/atomic"

	"github.com/FerNunez/tuto_go_server/internal/database"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
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
  godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbURL)
	dbQueries := database.New(db)

	cfg := apiConfig{}
	serverMux := http.NewServeMux()
	serverMux.Handle("/", http.StripPrefix("/app/", cfg.middlewareMetricsInc(http.FileServer(http.Dir(".")))))
	serverMux.Handle("assets/logo.png", http.FileServer(http.Dir("assets/logo.png")))
	serverMux.HandleFunc("GET /admin/metrics", cfg.metricsDisplayHandler)
	serverMux.HandleFunc("GET /api/healthz", readinnesHandler)
	serverMux.HandleFunc("POST /admin/reset", cfg.metricsResetHandler)
	serverMux.HandleFunc("POST /api/validate_chirp", validateChirps)

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
	rw.Header().Add("Content-Type", "text/html")
	rw.WriteHeader(200)
	x := int(cfg.fileserverHits.Load())
	fmt.Fprintf(rw, `<html>
											<body>
												<h1>Welcome, Chirpy Admin</h1>
												<p>Chirpy has been visited %d times!</p>
											</body>
										</html>`, x)
}

func (cfg *apiConfig) metricsResetHandler(rw http.ResponseWriter, _ *http.Request) {
	rw.Header().Add("Content-Type", "text/plain;charset=utf-8")
	rw.WriteHeader(200)

	cfg.fileserverHits.Store(0)
	rw.Write([]byte("Counter Reseted"))

}

func validateChirps(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")

	// Receice & Decode
	type Chirp struct {
		Body string `json:"body"`
	}

	type ChirpError struct {
		ErrResponse string `json:"error"`
	}

	type ValidResponse struct {
		CleanBody string `json:"cleaned_body"`
	}

	decoder := json.NewDecoder(r.Body)
	var chirp Chirp
	err := decoder.Decode(&chirp)
	if err != nil {
		fmt.Println("Error decoding chirp", err)
		errResp := ChirpError{"Something went wrong"}
		dat, _ := json.Marshal(errResp)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(dat)
		return
	}

	// Encode & Send
	if len(chirp.Body) > 10 {
		fmt.Println("Error decoding chirp", err)
		errResp := ChirpError{"Chirp is too long"}
		dat, _ := json.Marshal(errResp)
		w.WriteHeader(http.StatusBadRequest)
		w.Write(dat)
		return
	}

	validResp := ValidResponse{cleanProfane(chirp.Body)}
	dat, _ := json.Marshal(validResp)
	w.WriteHeader(http.StatusOK)
	w.Write(dat)

}

var bannedWords = []string{"kerfuffle", "sharbert", "fornax"}

func cleanProfane(input string) string {
	output := []string{}

	for _, word := range strings.Fields(input) {
		for _, banned := range bannedWords {
			if banned == strings.ToLower(word) {
				word = "****"
			}
		}
		output = append(output, word)
	}
	return strings.Join(output, " ")
}
