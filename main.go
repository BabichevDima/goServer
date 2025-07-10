package main

import (
	"fmt"
	"net/http"
)

func main() {
	fmt.Println("Server started on localhost:8080")
	mux := http.NewServeMux()

	mux.Handle("/app/", http.StripPrefix("/app/", http.FileServer(http.Dir("./"))))
	mux.Handle("/assets", http.FileServer(http.Dir("./")))
	// mux.HandleFunc("/healthz", healthzHandler)
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	err := server.ListenAndServe()
	if err != nil {
		panic(err)
	}
}

// func healthzHandler (w http.ResponseWriter, r *http.Request) {
// 		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
// 		w.WriteHeader(http.StatusOK)
// 		w.Write([]byte("OK"))
// }