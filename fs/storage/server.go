package storage

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func StartStorageServer() {
	server()
}

func server() {
	r := mux.NewRouter()

	r.HandleFunc("/chunks/{chunkId}", downloadChunk).Methods("GET")
	r.HandleFunc("/chunks/{chunkId}", uploadChunk).Methods("POST")

	server := http.Server{
		Handler: r,
		Addr:    ":8080",
	}

	log.Println("Starting storage HTTP server")

	log.Fatal(server.ListenAndServe())
}
