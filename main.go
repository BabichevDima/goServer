package main

import (
	"fmt"
	"net/http"
)

func main() {
	fmt.Println("Server started on localhost:8080")
	mux := http.NewServeMux()

	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	err := server.ListenAndServe()
	if err != nil {
		panic(err)
	}

	fmt.Println("Server started on localhost:8080")
}