package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	//"strings"
	"sync/atomic"
	"time"

	"github.com/FerNunez/tuto_go_server/internal/auth"
	"github.com/FerNunez/tuto_go_server/internal/database"
	"github.com/google/uuid"
)

type ApiConfig struct {
	FileserverHits atomic.Int32
	Db             *database.Queries
	Platform       string
}

func (cfg *ApiConfig) CreateUser(w http.ResponseWriter, r *http.Request) {

	type UserRequest struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	type CreatedResponse struct {
		Id        uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Email     string    `json:"email"`
	}

	decoder := json.NewDecoder(r.Body)

	var userReq UserRequest
	err := decoder.Decode(&userReq)
	if err != nil {
		errmsg := fmt.Sprintf("Could not decode incoming user data; %v", err)
		fmt.Println(errmsg)
		w.Header().Add("Content-Type", "text/plain;charset=utf-8")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(errmsg))
	}

	hashedPass, err := auth.HashPassword(userReq.Password)
	if err != nil {
		errmsg := fmt.Sprintf("Could not hash user password %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(errmsg))
		return
	}

	createUserParams := database.CreateUserParams{Email: userReq.Email, HashedPassword: hashedPass}
	user, errCreate := cfg.Db.CreateUser(r.Context(), createUserParams)
	if errCreate != nil {
		errmsg := fmt.Sprintf("Error creating response %v", errCreate)
		fmt.Println(errmsg)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(errmsg))
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
func (cfg *ApiConfig) GetChirpsHandler(w http.ResponseWriter, r *http.Request) {
	type ChirpResponse struct {
		Id        uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Body      string    `json:"body"`
		UserId    uuid.UUID `json:"user_id"`
	}

	// Todo: Understand why this needs a DB?
	chirps, err := cfg.Db.GetChirps(r.Context())
	if err != nil {
		fmt.Println("Error retrieving all chirps from DB: ", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	chirpsResponse := make([]ChirpResponse, len(chirps))
	for idx, chirpSql := range chirps {
		chirpsResponse[idx] = ChirpResponse{Id: chirpSql.ID, CreatedAt: chirpSql.CreatedAt, UpdatedAt: chirpSql.UpdatedAt, Body: chirpSql.Body, UserId: chirpSql.UserID}
	}
	dat, _ := json.Marshal(chirpsResponse)
	w.WriteHeader(http.StatusOK)
	w.Write(dat)
}

func (cfg *ApiConfig) GetChirpsByIDHandler(w http.ResponseWriter, r *http.Request) {

	type ChirpResponse struct {
		Id        uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Body      string    `json:"body"`
		UserId    uuid.UUID `json:"user_id"`
	}

	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		errmsg := fmt.Sprintf("Could not convert {id} key into chirp UUID type. Err: %v", err)
		w.Header().Add("Content-Type", "text/plain;charset=utf-8")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(errmsg))
		return
	}

	chirpSql, err := cfg.Db.GetChirpsByID(r.Context(), []uuid.UUID{id})
	if err != nil {
		errmsg := fmt.Sprintf("Chirp not found. Err: %v", err)
		println(errmsg)
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(errmsg))
		return
	}

	chirpResponse := ChirpResponse{Id: chirpSql.ID, CreatedAt: chirpSql.CreatedAt, UpdatedAt: chirpSql.UpdatedAt, Body: chirpSql.Body, UserId: chirpSql.UserID}

	dat, _ := json.Marshal(chirpResponse)
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(dat)
}

func (cfg *ApiConfig) Login(w http.ResponseWriter, r *http.Request) {

	type LoginReq struct {
		Password string `json:"password"`
		Email    string `json:"email"`
	}

	type LoginResp struct {
		Id        uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Email     string    `json:"email"`
	}

	// Unmarshal or Decode
	decoder := json.NewDecoder(r.Body)
	var loginReq LoginReq
	err := decoder.Decode(&loginReq)
	if err != nil {
		errmsg := fmt.Sprintf("Could not decode incoming login request. Err: %v", err)
		fmt.Println(errmsg)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(errmsg))
		return
	}

	dbUser, err := cfg.Db.GetUserByEmail(r.Context(), loginReq.Email)
	if err != nil {
		errmsg := fmt.Sprintf("Could not retrieve user. Err: %v", err)
		fmt.Println(errmsg)
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(errmsg))
		return
	}

	if err := auth.CheckPasswordHash(loginReq.Password, dbUser.HashedPassword); err != nil {
		errmsg := fmt.Sprintln("Wrong password")
		fmt.Println(errmsg)
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(errmsg))
		return
	}

	loginResp := LoginResp{Id: dbUser.ID, CreatedAt: dbUser.CreatedAt, UpdatedAt: dbUser.UpdatedAt, Email: dbUser.Email}
	dat, _ := json.Marshal(loginResp)
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(dat))
}
