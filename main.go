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
	"time"

	"github.com/FerNunez/tuto_go_server/internal/database"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	db             *database.Queries
	platform       string
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
	if err != nil {
		fmt.Println("error")
	}
	dbQueries := database.New(db)

	cfg := apiConfig{}
	cfg.db = dbQueries
	cfg.platform = os.Getenv("PLATFORM")

	serverMux := http.NewServeMux()
	serverMux.Handle("/", http.StripPrefix("/app/", cfg.middlewareMetricsInc(http.FileServer(http.Dir(".")))))
	serverMux.Handle("assets/logo.png", http.FileServer(http.Dir("assets/logo.png")))
	serverMux.HandleFunc("GET /admin/metrics", cfg.metricsDisplayHandler)
	serverMux.HandleFunc("GET /api/healthz", readinnesHandler)
	serverMux.HandleFunc("POST /admin/reset", cfg.metricsResetHandler)
	serverMux.HandleFunc("POST /api/user", cfg.createUser)
	serverMux.HandleFunc("POST /api/chirps", cfg.createChirp)

	// Listen & Serve
	server := &http.Server{
		Handler: serverMux,
		Addr:    ":8080",
	}
	server.ListenAndServe()

}

func (cfg *apiConfig) createUser(w http.ResponseWriter, r *http.Request) {

	type emailRequest struct {
		Email string `json:"email"`
	}

	type CreatedResponse struct {
		Id        uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Email     string    `json:"email"`
	}

	decoder := json.NewDecoder(r.Body)

	var emailReq emailRequest
	err := decoder.Decode(&emailReq)
	if err != nil {
		fmt.Println("Error email request decoding")
		w.Header().Add("Content-Type", "text/plain;charset=utf-8")
		w.WriteHeader(http.StatusInternalServerError)
	}

	user, errCreate := cfg.db.CreateUser(r.Context(), emailReq.Email)
	fmt.Println("Trying to add", emailReq.Email)
	if errCreate != nil {
		fmt.Println("Error creating response", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	createdRespo := CreatedResponse{Id: user.ID, CreatedAt: user.CreatedAt, UpdatedAt: user.UpdatedAt, Email: user.Email}
	dat, err := json.Marshal(createdRespo)
	if err != nil {
		fmt.Println("Error creating response")
		w.WriteHeader(http.StatusInternalServerError)
	}
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(201)
	w.Write(dat)
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
	if cfg.platform != "dev" {
		rw.Header().Add("Content-Type", "text/plain;charset=utf-8")
		rw.WriteHeader(403)
	}

	rw.Header().Add("Content-Type", "text/plain;charset=utf-8")
	rw.WriteHeader(200)

	cfg.fileserverHits.Store(0)
	dbURL := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		fmt.Println("error")
	}
	cfg.db = database.New(db)
	rw.Write([]byte("Counter and DB Reseted "))

}

func (cfg *apiConfig) createChirp(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")

	// Receice & Decode
	type ChirpReq struct {
		Body   string    `json:"body"`
		UserId uuid.UUID `json:"user_id"`
	}

	type ChirpError struct {
		ErrResponse string `json:"error"`
	}

	type ChirpResponse struct {
		Id        uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Body      string    `json:"body"`
		UserId    uuid.UUID `json:"user_id"`
	}

	decoder := json.NewDecoder(r.Body)
	var chirp ChirpReq
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

	chirpArgs := database.CreateChirpParams{Body: chirp.Body, UserID: chirp.UserId}
	chirpy, err1 := cfg.db.CreateChirp(r.Context(), chirpArgs)
	if err1 != nil {
		fmt.Println("Error saving in DB err: ", err1)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	chirpyResp := ChirpResponse{Id: chirpy.ID, CreatedAt: chirpy.CreatedAt, UpdatedAt: chirpy.UpdatedAt, Body: chirpy.Body, UserId: chirpy.UserID}
	dat, _ := json.Marshal(chirpyResp)
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
