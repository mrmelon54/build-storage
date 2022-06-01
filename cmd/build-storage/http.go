package main

import (
	"github.com/gorilla/mux"
	"net/http"
	"os"
)

func setupHttpServer() *http.Server {
	router := mux.NewRouter()
	router.HandleFunc("/", func(rw http.ResponseWriter, req *http.Request) {
		http.Error(rw, "Hi", http.StatusOK)
	})

	httpServer := &http.Server{
		Addr:    os.Getenv("LISTEN"),
		Handler: router,
	}
	return httpServer
}
