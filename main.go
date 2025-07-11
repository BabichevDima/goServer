// Package main implements a web server with hit counter metrics.
// It serves static files from /app/ and /assets/ endpoints,
// and provides API endpoints for health checks and metrics.
package main

import (
	"fmt"
	"net/http"
	"sync/atomic"
	"log"
	"html/template"
    "encoding/json"
)

// apiConfig holds application configuration and shared state.
// The fileserverHits field tracks the number of requests made to the fileserver.
type apiConfig struct {
	fileserverHits atomic.Int32
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

// main initializes and starts the HTTP server on localhost:8080.
// It sets up routes for:
// - /app/ (file server with hit tracking)
// - /assets/ (static file server)
// - /api/healthz (health check endpoint)
// - /api/metrics (hit counter metrics)
// - /api/reset (hit counter reset)
func main() {
	fmt.Println("Server started on localhost:8080")
	mux := http.NewServeMux()
	apiCfg := &apiConfig{}

	// Fileservers
	mux.Handle("/app/", apiCfg.middlewareMetricsInc(http.StripPrefix("/app/", http.FileServer(http.Dir("./")))))
	mux.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("./"))))

	// API endpoints
	mux.Handle("GET /api/healthz", middlewareLog(http.HandlerFunc(healthzHandler)))
	mux.Handle("POST /api/validate_chirp", middlewareLog(http.HandlerFunc(handlerValidateChirp)))
	mux.Handle("POST /admin/reset", middlewareLog(http.HandlerFunc(apiCfg.handlerReset)))
	mux.Handle("GET /admin/metrics", middlewareLog(http.HandlerFunc(apiCfg.handlerMetrics)))
	// mux.HandleFunc("GET /api/healthz", healthzHandler)
	// mux.HandleFunc("GET /api/metrics", apiCfg.handlerMetrics)
	// mux.HandleFunc("POST /api/reset", apiCfg.handlerReset)

	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	err := server.ListenAndServe()

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

// handlerValidateChirp handles POST requests to validate Chirps.
// It expects a JSON body with a "body" field containing the chirp text.
//
// Valid chirps must:
// - Exist (non-empty)
// - Be 140 characters or less
//
// Responses:
//   - 200 OK with {"valid":true} for valid chirps
//   - 400 Bad Request with {"error":"message"} for invalid chirps or bad requests
func handlerValidateChirp (w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body string `json:"body"`
	}
	type returnVals struct {
		Valid bool `json:"valid"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)

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

	// respondWithJSON(w, http.StatusOK, returnVals{Valid: true})
	
	respBody := returnVals{
		Valid: true,
	}
	dat, err := json.Marshal(respBody)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, fmt.Sprintf("Something went wrong. Error marshalling JSON: %s", err))
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write(dat)
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