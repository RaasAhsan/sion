package metadata

import (
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

func StartMetadataProcess() {
	server()
}

type MetadataHandler struct{}

func (h *MetadataHandler) Join(w http.ResponseWriter, r *http.Request) {
	id := uuid.New()
	w.Write([]byte(id.String()))
}

func (h *MetadataHandler) Heartbeat(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("ok"))
}

func server() {
	r := mux.NewRouter()

	handler := &MetadataHandler{}

	r.HandleFunc("/join", handler.Join).Methods("POST")
	r.HandleFunc("/heartbeat", handler.Heartbeat).Methods("POST")

	server := http.Server{
		Handler: r,
		Addr:    ":8000",
	}

	log.Println("Starting metadata HTTP server")

	log.Fatal(server.ListenAndServe())
}
