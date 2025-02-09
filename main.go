// The Go standard library makes it easy to build a simple server. Your task is to build and run a server that binds to localhost:8080 and always responds with a 404 Not Found response.
package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"

	"github.com/FerNunez/tuto_go_server/internal/api"
	"github.com/FerNunez/tuto_go_server/internal/database"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		fmt.Println("error")
	}
	dbQueries := database.New(db)

	cfg := api.ApiConfig{Db: dbQueries, Platform: os.Getenv("PLATFORM"), SignString: os.Getenv("SIGN_STRING")}

	serverMux := http.NewServeMux()
	serverMux.Handle("/", http.StripPrefix("/app/", cfg.MiddlewareMetricsInc(http.FileServer(http.Dir(".")))))
	serverMux.Handle("assets/logo.png", http.FileServer(http.Dir("assets/logo.png")))
	serverMux.HandleFunc("GET /admin/metrics", cfg.MetricsDisplayHandler)
	serverMux.HandleFunc("GET /api/healthz", api.ReadinnesHandler)
	serverMux.HandleFunc("POST /admin/reset", cfg.ResetHandler)
	serverMux.HandleFunc("POST /api/user", cfg.CreateUser)
	serverMux.HandleFunc("PUT /api/user", cfg.UpdateUserHandler)
	serverMux.HandleFunc("POST /api/chirps", cfg.CreateChirp)
	serverMux.HandleFunc("GET /api/chirps", cfg.GetChirpsHandler)
	serverMux.HandleFunc("GET /api/chirps/{id}", cfg.GetChirpsByIDHandler)
	serverMux.HandleFunc("POST /api/login", cfg.LoginHandler)
	serverMux.HandleFunc("POST /api/refresh", cfg.RefreshHandler)
	serverMux.HandleFunc("POST /api/revoke", cfg.RevokeHandler)


	// Listen & Serve
	server := &http.Server{
		Handler: serverMux,
		Addr:    ":8080",
	}
	server.ListenAndServe()

}
