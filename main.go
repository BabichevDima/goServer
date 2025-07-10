package main

import (
	"fmt"
	"net/http"
	"sync/atomic"
	"log"
)

type apiConfig struct {
	fileserverHits atomic.Int32
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) handlerMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	hits := cfg.fileserverHits.Load()
	w.Write([]byte(fmt.Sprintf("Hits: %d", hits)))
}

func (cfg *apiConfig) handlerReset(w http.ResponseWriter, r *http.Request) {
	cfg.fileserverHits.Store(0)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Hits reset to 0"))
}

func main() {
	fmt.Println("Server started on localhost:8080")
	mux := http.NewServeMux()
	apiCfg := &apiConfig{}

	// MetricsInc
	fileServer := http.StripPrefix("/app/", http.FileServer(http.Dir("./")))
	mux.Handle("/app/", apiCfg.middlewareMetricsInc(fileServer))

	// Not MetricsInc
	mux.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("./"))))

	mux.HandleFunc("GET /healthz", healthzHandler)
	mux.HandleFunc("GET /metrics", apiCfg.handlerMetrics)
	// mux.Handle("GET /metrics", middlewareLog(http.HandlerFunc(apiCfg.handlerMetrics)))
	mux.HandleFunc("POST /reset", apiCfg.handlerReset)

	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	err := server.ListenAndServe()

	if err != nil {
		panic(err)
	}
}

func middlewareLog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s", r.Method, r.URL.Path)
		// next.ServeHTTP(w, r)
	})
}

func healthzHandler (w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
}