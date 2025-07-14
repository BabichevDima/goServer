// Package main implements a web server with hit counter metrics.
// It serves static files from /app/ and /assets/ endpoints,
// and provides API endpoints for health checks and metrics.
package main

import (
	_ "github.com/lib/pq"
	"github.com/joho/godotenv"

	"database/sql"
	"os"
	"fmt"
	"net/http"
	"sync/atomic"
	"log"
	"html/template"
    "encoding/json"
	"regexp"
	"time"
	"strings"
	"errors"
	
	"github.com/BabichevDima/goServer/internal/database"
	"github.com/BabichevDima/goServer/internal/auth"
	"github.com/google/uuid"
)

// apiConfig holds application configuration and shared state.
// The fileserverHits field tracks the number of requests made to the fileserver.
type apiConfig struct {
	jwtSecret		string
	fileserverHits	atomic.Int32
	DB				*database.Queries 
}

type User struct {
	ID           string    `json:"id"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	Email        string    `json:"email"`
	Token        string    `json:"token,omitempty"`
	RefreshToken string    `json:"refresh_token,omitempty"`
}

type Chirp struct {
    ID        uuid.UUID `json:"id"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
    Body      string    `json:"body"`
    UserID    uuid.UUID `json:"user_id"`
}

// middlewareMetricsInc creates a middleware that increments the hit counter
// for each request before calling the next handler.
func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

// handlerMetrics writes the current hit count in the format "Hits: N".
// It responds with Content-Type: text/html and HTTP 200 status.
func (cfg *apiConfig) handlerMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	hits := cfg.fileserverHits.Load()
	tmpl := template.Must(template.New("metrics").Parse(`
	<html>
	<body>
		<h1>Welcome, Chirpy Admin</h1>
		<p>Chirpy has been visited {{.}} times!</p>
	</body>
	</html>
	`))
    
    tmpl.Execute(w, hits)

}

// handlerReset sets the hit counter back to zero.
// It responds with "Hits reset to 0", Content-Type: text/plain and HTTP 200 status.
func (cfg *apiConfig) handlerReset(w http.ResponseWriter, r *http.Request) {
	cfg.fileserverHits.Store(0)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Hits reset to 0"))
}

// connectToDB establishes a connection to the PostgreSQL database using the connection URL
// from the environment variable DB_URL (loaded via .env file).
//
// It returns:
//   - *database.Queries: A prepared queries object for database operations
//   - error: Any error that occurred during connection (e.g., environment loading failure,
//     invalid connection URL, or connection failure)
//
// Example usage:
//   queries, err := connectToDB()
//   if err != nil {
//       log.Fatal(err)
//   }
//   defer queries.Close()
func connectToBD() (*database.Queries, string, error) {
	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	jwtSecret := os.Getenv("JWT_SECRET")
	fmt.Println("dbURL = ", dbURL)

	//  sql.Open() a connection to your database
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		return nil, "", fmt.Errorf("failed to connect to db: %w", err)
	}

	dbQueries := database.New(db)

	return dbQueries, jwtSecret, nil
}

// main initializes and starts the HTTP server on localhost:8080.
// It sets up routes for:
// - /app/ (file server with hit tracking)
// - /assets/ (static file server)
// - /api/healthz (health check endpoint)
// - /api/metrics (hit counter metrics)
// - /api/reset (hit counter reset)
func main() {
	dbQueries, jwtSecret, err := connectToBD()
	// fmt.Println("dbQueries = ", dbQueries)

	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	fmt.Println("Server started on localhost:8080")
	mux := http.NewServeMux()
	apiCfg := &apiConfig{
		DB: dbQueries,
		jwtSecret: jwtSecret,
	}

	if apiCfg.jwtSecret == "" {
		log.Fatal("JWT_SECRET environment variable is not set")
	}

	// Fileservers
	mux.Handle("/app/", apiCfg.middlewareMetricsInc(http.StripPrefix("/app/", http.FileServer(http.Dir("./")))))
	mux.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("./"))))

	// API endpoints
	mux.Handle("GET /api/healthz", middlewareLog(http.HandlerFunc(healthzHandler)))
	mux.Handle("POST /api/users", middlewareLog(http.HandlerFunc(apiCfg.handlerCreateUser)))
	mux.Handle("POST /api/login", middlewareLog(http.HandlerFunc(apiCfg.handlerLogin)))
	mux.Handle("POST /api/refresh", middlewareLog(http.HandlerFunc(apiCfg.handlerRefresh)))
	mux.Handle("POST /api/revoke", middlewareLog(http.HandlerFunc(apiCfg.handlerRevoke)))
	mux.Handle("POST /api/chirps", middlewareLog(http.HandlerFunc(apiCfg.handlerCreateChirp)))
	mux.Handle("GET /api/chirps", middlewareLog(http.HandlerFunc(apiCfg.handlerGetChirps)))
	mux.Handle("GET /api/chirps/{chirpID}", middlewareLog(http.HandlerFunc(apiCfg.handlerGetChirp)))

	mux.Handle("POST /admin/reset", middlewareLog(http.HandlerFunc(apiCfg.handlerReset)))
	mux.Handle("GET /admin/metrics", middlewareLog(http.HandlerFunc(apiCfg.handlerMetrics)))

	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	err = server.ListenAndServe()

	if err != nil {
		panic(err)
	}
}

// middlewareLog creates a middleware that logs the HTTP method and path
// of each request before calling the next handler.
func middlewareLog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

// healthzHandler responds to health check requests.
// It always returns "OK" with Content-Type: text/plain and HTTP 200 status.
func healthzHandler (w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
}

func (cfg *apiConfig) handlerCreateUser(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email string `json:"email"`
		Password string `json:"password"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)

	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if params.Email == "" {
		respondWithError(w, http.StatusBadRequest, "Email is required")
		return
	}
	if params.Password == "" {
		respondWithError(w, http.StatusBadRequest, "Password is required")
		return
	}

	hashPassword, err := auth.HashPassword(params.Password)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Email HashPassword")
		return
	}
	user, err := cfg.DB.CreateUser(r.Context(), database.CreateUserParams{
		Email:			params.Email,
		HashedPassword:	hashPassword,
	})
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			respondWithError(w, http.StatusConflict, "Email already exists")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Failed to create user")
		return
	}

	respondWithJSON(w, http.StatusCreated, User{
		ID:        user.ID.String(),
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     user.Email,
	})
}

func (cfg *apiConfig) handlerLogin(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email string `json:"email"`
		Password string `json:"password"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if params.Email == "" || params.Password == "" {
		respondWithError(w, http.StatusBadRequest, "Email and password are required")
		return
	}

	user, err := cfg.DB.GetUserByEmail(r.Context(), params.Email)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to get user")
		return
	}

	if err := auth.CheckPasswordHash(params.Password, user.HashedPassword); err != nil {
		respondWithError(w, http.StatusUnauthorized, "Incorrect email or password")
		return
	}

	accessToken, err := auth.MakeJWT(user.ID, cfg.jwtSecret, time.Hour)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to create access Token")
		return
	}

	refreshToken, err := auth.MakeRefreshToken()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to create Refresh Token")
		return
	}

	expiresAt := time.Now().Add(60 * 24 * time.Hour)
	_, err = cfg.DB.CreateRefreshToken(r.Context(), database.CreateRefreshTokenParams{
		Token:     refreshToken,
		UserID:    user.ID,
		ExpiresAt: expiresAt,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to save refresh token")
		return
	}

	respondWithJSON(w, http.StatusOK, User{
		ID:           user.ID.String(),
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
		Email:        user.Email,
		Token:        accessToken,
		RefreshToken: refreshToken,
	})
}

func (cfg *apiConfig) handlerRefresh(w http.ResponseWriter, r *http.Request) {
	refreshToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid authorization header")
		return
	}

	user, err := cfg.DB.GetUserFromRefreshToken(r.Context(), refreshToken)
	fmt.Println(user.ID)
	fmt.Println(user.Email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respondWithError(w, http.StatusUnauthorized, "Invalid or expired refresh token")
		} else {
			respondWithError(w, http.StatusInternalServerError, "Failed to validate refresh token")
		}
		return
	}

	accessToken, err := auth.MakeJWT(user.ID, cfg.jwtSecret, time.Hour)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to generate access token")
		return
	}

	respondWithJSON(w, http.StatusOK, struct {
		Token string `json:"token"`
	}{
		Token: accessToken,
	})
}

func (cfg *apiConfig) handlerRevoke(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	if err := cfg.DB.RevokeRefreshToken(r.Context(), token); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to Refresh Token")
		return
	}

	respondWithJSON(w, http.StatusNoContent, nil)
}

func (cfg *apiConfig) handlerCreateChirp(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body string `json:"body"`
	}

	token, err := auth.GetBearerToken(r.Header)
    if err != nil {
        respondWithError(w, http.StatusUnauthorized, "Authentication required")
        return
    }

	userID, err := auth.ValidateJWT(token, cfg.jwtSecret)
    if err != nil {
        respondWithError(w, http.StatusUnauthorized, "Invalid token")
        return
    }

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err = decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Something went wrong")
		return
	}

	if params.Body == "" {
		respondWithError(w, http.StatusBadRequest, "Chirp is empty")
		return
	} else if len(params.Body) > 140 {
		respondWithError(w, http.StatusBadRequest, "Chirp is too long")
		return
	}

	cleanedBody := replacer(params.Body)

	chirp, err := cfg.DB.CreateChirp(r.Context(), database.CreateChirpParams{
		Body:   cleanedBody,
		UserID: userID,
	})
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			respondWithError(w, http.StatusConflict, "Chirp already exists")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Failed to create chirp")
		return
	}
	respondWithJSON(w, http.StatusCreated, Chirp{
		ID:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserID:    chirp.UserID,
	})
}

func (cfg *apiConfig) handlerGetChirps(w http.ResponseWriter, r *http.Request) {
	dbChirps, err := cfg.DB.GetChirps(r.Context())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to get chirps")
		return
	}

	chirps := make([]Chirp, len(dbChirps))

	for i, dbChirp := range dbChirps {
		chirps[i] = Chirp{
			ID:        dbChirp.ID,
			CreatedAt: dbChirp.CreatedAt,
			UpdatedAt: dbChirp.UpdatedAt,
			Body:      dbChirp.Body,
			UserID:    dbChirp.UserID,
		}
	}

	respondWithJSON(w, http.StatusOK, chirps)
}

func (cfg *apiConfig) handlerGetChirp(w http.ResponseWriter, r *http.Request) {
	chirpIDStr := r.PathValue("chirpID")

	chirpID, err := uuid.Parse(chirpIDStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid chirpID format")
		return
	}

	dbChirp, err := cfg.DB.GetChirp(r.Context(), chirpID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to get chirp")
		return
	}

	respondWithJSON(w, http.StatusOK, Chirp{
		ID:        dbChirp.ID,
		CreatedAt: dbChirp.CreatedAt,
		UpdatedAt: dbChirp.UpdatedAt,
		Body:      dbChirp.Body,
		UserID:    dbChirp.UserID,
	})
}

// respondWithError is a helper function to send JSON error responses.
// It sets the appropriate content type and HTTP status code,
// then encodes the error message as JSON.
func respondWithError(w http.ResponseWriter, code int, message string) {
    respondWithJSON(w, code, map[string]string{"error": message})
}

// respondWithJSON is a helper function to send JSON responses.
// It sets the Content-Type header to application/json,
// writes the HTTP status code, and encodes the payload as JSON.
func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(payload)
}

// replacer replaces banned words with asterisks
// It takes a string input and returns a sanitized version
// where "kerfuffle", "sharbert" and "fornax" are replaced with "****"
func replacer (str string) string {
	patterns := map[string]*regexp.Regexp{
		"kerfuffle": regexp.MustCompile(`(?i)kerfuffle`),
		"sharbert":  regexp.MustCompile(`(?i)sharbert`),
		"fornax":    regexp.MustCompile(`(?i)fornax`),
	}

	for _, re := range patterns {
		str = re.ReplaceAllString(str, "****")
	}
	return str
}