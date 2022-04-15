package metadata

import (
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

func StartMetadataServer() {
	server()
}

func Join(w http.ResponseWriter, r *http.Request) {
	id := uuid.New()
	w.Write([]byte(id.String()))
}

func Heartbeat(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("ok"))
}

func server() {
	r := mux.NewRouter()

	r.HandleFunc("/join", Join).Methods("POST")
	r.HandleFunc("/heartbeat", Heartbeat).Methods("POST")

	server := http.Server{
		Handler: r,
		Addr:    ":8000",
	}

	log.Println("Starting metadata HTTP server")

	log.Fatal(server.ListenAndServe())
}
