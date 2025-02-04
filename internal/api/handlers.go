package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync/atomic"
	"time"

	"github.com/FerNunez/tuto_go_server/internal/database"
	"github.com/google/uuid"
)

type ApiConfig struct {
	FileserverHits atomic.Int32
	Db             *database.Queries
	Platform       string
}

func (cfg *ApiConfig) CreateUser(w http.ResponseWriter, r *http.Request) {

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

	user, errCreate := cfg.Db.CreateUser(r.Context(), emailReq.Email)
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

func ReadinnesHandler(rw http.ResponseWriter, _ *http.Request) {
	rw.Header().Add("Content-Type", "text/plain;charset=utf-8")
	rw.WriteHeader(200)
	rw.Write([]byte("Ok"))
}

func (cfg *ApiConfig) MetricsDisplayHandler(rw http.ResponseWriter, _ *http.Request) {
	rw.Header().Add("Content-Type", "text/html")
	rw.WriteHeader(200)
	x := int(cfg.FileserverHits.Load())
	fmt.Fprintf(rw, `<html>
											<body>
												<h1>Welcome, Chirpy Admin</h1>
												<p>Chirpy has been visited %d times!</p>
											</body>
										</html>`, x)
}

func (cfg *ApiConfig) MetricsResetHandler(rw http.ResponseWriter, _ *http.Request) {
	if cfg.Platform != "dev" {
		rw.Header().Add("Content-Type", "text/plain;charset=utf-8")
		rw.WriteHeader(403)
	}

	rw.Header().Add("Content-Type", "text/plain;charset=utf-8")
	rw.WriteHeader(200)

	cfg.FileserverHits.Store(0)
	dbURL := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		fmt.Println("error")
	}
	cfg.Db = database.New(db)
	rw.Write([]byte("Counter and DB Reseted "))

}

func (cfg *ApiConfig) CreateChirp(w http.ResponseWriter, r *http.Request) {
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
	chirpy, err1 := cfg.Db.CreateChirp(r.Context(), chirpArgs)
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

