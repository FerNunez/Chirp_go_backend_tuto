package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
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
	SignString     string
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

func (cfg *ApiConfig) ResetHandler(rw http.ResponseWriter, r *http.Request) {
	if cfg.Platform != "dev" {
		rw.Header().Add("Content-Type", "text/plain;charset=utf-8")
		rw.WriteHeader(403)
		return
	}

	cfg.FileserverHits.Store(0)

	// Reset db
	cfg.Db.ResetUsers(r.Context())

	rw.Header().Add("Content-Type", "text/plain;charset=utf-8")
	rw.WriteHeader(200)
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

	// Get Header Authentification
	authToken, errAuth := auth.GetBearerToken(r.Header)
	if errAuth != nil {
		errmsg := fmt.Sprintf("could not get header authorization: %v", errAuth)
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(errmsg))
		return
	}

	// Decode input
	decoder := json.NewDecoder(r.Body)
	var chirpReq ChirpReq
	err := decoder.Decode(&chirpReq)
	if err != nil {
		fmt.Println("Error decoding chirp", err)
		errResp := ChirpError{"Something went wrong"}
		dat, _ := json.Marshal(errResp)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(dat)
		return
	}

	// Encode & Send
	if len(chirpReq.Body) > 100 {
		fmt.Println("Chirp length overpassed", err)
		errResp := ChirpError{"Chirp is too long"}
		dat, _ := json.Marshal(errResp)
		w.WriteHeader(http.StatusBadRequest)
		w.Write(dat)
		return
	}

	// Validate user by Authentification Token
	validatedUserId, err := auth.ValidateJWT(authToken, cfg.SignString)
	if err != nil || validatedUserId != chirpReq.UserId {
		errmsg := fmt.Sprintf("user not authenticated: %v", err)
		fmt.Println(errmsg)
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(errmsg))
		return
	}

	chirpArgs := database.CreateChirpParams{Body: chirpReq.Body, UserID: chirpReq.UserId}
	chirpy, err1 := cfg.Db.CreateChirp(r.Context(), chirpArgs)
	if err1 != nil {
		fmt.Println("Error saving in DB err: ", err1)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	chirpyResp := ChirpResponse{Id: chirpy.ID, CreatedAt: chirpy.CreatedAt, UpdatedAt: chirpy.UpdatedAt, Body: chirpy.Body, UserId: chirpy.UserID}
	dat, _ := json.Marshal(chirpyResp)
	w.WriteHeader(http.StatusCreated)
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

func (cfg *ApiConfig) LoginHandler(w http.ResponseWriter, r *http.Request) {

	type LoginReq struct {
		Password string `json:"password"`
		Email    string `json:"email"`
	}

	type LoginResp struct {
		Id           uuid.UUID `json:"id"`
		CreatedAt    time.Time `json:"created_at"`
		UpdatedAt    time.Time `json:"updated_at"`
		Email        string    `json:"email"`
		Token        string    `json:"token"`
		RefreshToken string    `json:"refresh_token"`
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

	// Get user in DB
	dbUser, err := cfg.Db.GetUserByEmail(r.Context(), loginReq.Email)
	if err != nil {
		errmsg := fmt.Sprintf("Could not retrieve user. Err: %v", err)
		fmt.Println(errmsg)
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(errmsg))
		return
	}

	// Check Password
	if err := auth.CheckPasswordHash(loginReq.Password, dbUser.HashedPassword); err != nil {
		errmsg := fmt.Sprintln("Wrong password")
		fmt.Println(errmsg)
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(errmsg))
		return
	}

	// JWToken creation
	jwtToken, err := auth.MakeJWT(dbUser.ID, cfg.SignString, time.Duration(3600)*time.Second)
	if err != nil {
		errmsg := fmt.Sprintf("Could not make token . Err: %v", err)
		fmt.Println(errmsg)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(errmsg))
		return
	}

	// RefreshToken
	refreshToken, errRefTok := auth.MakeRefreshToken()
	if errRefTok != nil {
		errmsg := fmt.Sprintf("Could not generate refresh token %v", err)
		fmt.Println(errmsg)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(errmsg))
	}
	createRefreshTokenParams := database.CreateRefreshTokenParams{Token: refreshToken, UserID: dbUser.ID, ExpiresAt: time.Now().UTC().Add(time.Hour * 24 * 60)}
	cfg.Db.CreateRefreshToken(r.Context(), createRefreshTokenParams)

	// Response
	loginResp := LoginResp{Id: dbUser.ID, CreatedAt: dbUser.CreatedAt, UpdatedAt: dbUser.UpdatedAt, Email: dbUser.Email, Token: jwtToken, RefreshToken: refreshToken}
	dat, _ := json.Marshal(loginResp)
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(dat))
}

func (cfg *ApiConfig) RefreshHandler(w http.ResponseWriter, r *http.Request) {

	// do not accept a body in requet
	if r.ContentLength > 0 {
		errmsg := fmt.Sprintln("request not allowed to contain a body")
		fmt.Println(errmsg)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(errmsg))
		return
	}

	// Check Authoritzon
	authToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		errmsg := fmt.Sprintf("could not retrieve authorization token: %v", err)
		fmt.Println(errmsg)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(errmsg))
		return
	}

	tokenFromDb, err := cfg.Db.GetRefreshToken(r.Context(), authToken)
	if err != nil {
		errmsg := fmt.Sprintf("could not retireve authorization from DB: %v", err)
		fmt.Println(errmsg)
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(errmsg))
		return
	}

	if tokenFromDb.ExpiresAt.UTC().Before(time.Now().UTC()) || (tokenFromDb.RevokedAt.Valid && tokenFromDb.RevokedAt.Time.UTC().Before(time.Now().UTC())) {
		errmsg := fmt.Sprintln("token expired or revoked")
		fmt.Println(errmsg)
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(errmsg))
		return
	}

	// Respond: new the jwt token
	jwtToken, err := auth.MakeJWT(tokenFromDb.UserID, cfg.SignString, time.Hour)
	if err != nil {
		errmsg := fmt.Sprintf("could not create new jwtToken: %v", err)
		fmt.Println(errmsg)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(errmsg))
		return
	}

	type tokenResponse struct {
		Token string `json:"token"`
	}

	dat, err := json.Marshal(tokenResponse{Token: jwtToken})
	if err != nil {
		errmsg := fmt.Sprintf("could not respond token: %v", err)
		fmt.Println(errmsg)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(errmsg))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(dat))
}

func (cfg *ApiConfig) RevokeHandler(w http.ResponseWriter, r *http.Request) {
	if r.ContentLength > 0 {
		errmsg := fmt.Sprintln("request not allowed to contain body")
		fmt.Println(errmsg)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(errmsg))
		return
	}

	// Get Refresh token
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		errmsg := fmt.Sprintf("could not retrieve token in request: %v", err)
		fmt.Println(errmsg)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(errmsg))
		return
	}
	dbRefreshToken, err := cfg.Db.GetRefreshToken(r.Context(), token)
	if err != nil {
		errmsg := fmt.Sprintf("could not find refresh token in db: %v", err)
		fmt.Println(errmsg)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(errmsg))
		return
	}

	// Revoking in DB
	revokedAt := sql.NullTime{Valid: true, Time: time.Now().UTC()}
	updateRefreshTokenParams := database.UpdateRefreshTokenParams{RevokedAt: revokedAt, UpdatedAt: time.Now().UTC(), Token: dbRefreshToken.Token}
	errRevoke := cfg.Db.UpdateRefreshToken(r.Context(), updateRefreshTokenParams)
	if errRevoke != nil {
		errmsg := fmt.Sprintf("could not revoke token: %v", errRevoke)
		fmt.Println(errmsg)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(errmsg))
		return
	}
	w.WriteHeader(http.StatusNoContent)

}

func (cfg *ApiConfig) UpdateUserHandler(w http.ResponseWriter, r *http.Request) {

	type UpdateUserReq struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	// Decode body
	var updateUserReq UpdateUserReq
	errDecode := json.NewDecoder(r.Body).Decode(&updateUserReq)
	if errDecode != nil {
		errmsg := fmt.Sprintf("could not decode request body: %v", errDecode)
		fmt.Println(errmsg)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(errmsg))
		return
	}

	// Validate access toke
	accessToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		errmsg := fmt.Sprintf("could not retrieve access token: %v", err)
		fmt.Println(errmsg)
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(errmsg))
		return
	}
	userID, err := auth.ValidateJWT(accessToken, cfg.SignString)
	if err != nil {
		errmsg := fmt.Sprintf("could not validate token: %v", err)
		fmt.Println(errmsg)
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(errmsg))
		return
	}

	// Hash Password
	hashedPassword, err := auth.HashPassword(updateUserReq.Password)
	if err != nil {
		errmsg := fmt.Sprintf("could not hash password: %v", err)
		fmt.Println(errmsg)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(errmsg))
		return
	}

	// Update Pass
	updateUserLoginByIDParams := database.UpdateUserLoginByIDParams{
		Email:          updateUserReq.Email,
		HashedPassword: hashedPassword,
		ID:             userID,
	}
	errUpdate := cfg.Db.UpdateUserLoginByID(r.Context(), updateUserLoginByIDParams)
	if errUpdate != nil {
		errmsg := fmt.Sprintf("could not update email&password: %v", errUpdate)
		fmt.Println(errmsg)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(errmsg))
		return
	}

	// Respond
	type UpdateUserResp struct {
		Id    uuid.UUID `json:"id"`
		Email string    `json:"email"`
	}
	updateUserResp := UpdateUserResp{
		Id:    userID,
		Email: updateUserReq.Email,
	}
	dat, err := json.Marshal(updateUserResp)
	if err != nil {
		errmsg := fmt.Sprintf("could not encode response: %v", err)
		fmt.Println(errmsg)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(errmsg))
		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(dat))

}
