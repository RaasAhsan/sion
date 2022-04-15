package metadata

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func StartMetadataServer() {
	server()
}

func heartbeat(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("ok"))
}

func server() {
	r := mux.NewRouter()

	// r.HandleFunc("/register", downloadChunk).Methods("POST")
	r.HandleFunc("/heartbeat", heartbeat).Methods("POST")

	server := http.Server{
		Handler: r,
		Addr:    ":8000",
	}

	log.Println("Starting metadata HTTP server")

	log.Fatal(server.ListenAndServe())
}
