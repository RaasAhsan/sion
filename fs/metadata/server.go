package metadata

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

func StartMetadataProcess() {
	server()
}

type MetadataHandler struct {
	// TODO: a lock for each subsystem?
	lock      sync.RWMutex
	namespace *Namespace
	cluster   *Cluster
}

func (h *MetadataHandler) Join(w http.ResponseWriter, r *http.Request) {
	id := uuid.New()
	w.Write([]byte(id.String()))
}

func (h *MetadataHandler) Heartbeat(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("ok"))
}

// TODO: Locate this business logic to another file, and just parse request/render response here?

func (h *MetadataHandler) GetFile(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	path := Path(params["path"])

	// TODO: scope this in a function to minimize critical region?
	h.lock.RLock()
	defer h.lock.RUnlock()

	if !h.namespace.FileExists(path) {
		http.Error(w, "File does not exist", http.StatusNotFound)
		return
	}

	file := h.namespace.GetFile(path)

	json, err := json.MarshalIndent(file, "", "  ")
	if err != nil {
		http.Error(w, "Failed to return response", http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.Write(json)
}

func (h *MetadataHandler) CreateFile(w http.ResponseWriter, r *http.Request) {
	// TODO: just inline this to CreateFile?
	type CreateFile struct {
		Path Path
	}
	var create CreateFile
	err := json.NewDecoder(r.Body).Decode(&create)
	if err != nil {
		http.Error(w, "Failed to parse body", http.StatusBadRequest)
		return
	}

	// TODO: scope this in a function to minimize critical region?
	h.lock.Lock()
	defer h.lock.Unlock()

	if h.namespace.FileExists(create.Path) {
		http.Error(w, "File already exists", http.StatusBadRequest)
		return
	}

	file := NewFile(create.Path)
	h.namespace.AddFile(file)

	// TODO: Separate API response type
	json, err := json.MarshalIndent(file, "", "  ")
	if err != nil {
		http.Error(w, "Failed to return response", http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.Write(json)
}

func server() {
	r := mux.NewRouter()

	handler := &MetadataHandler{namespace: NewNamespace(), cluster: nil}

	r.HandleFunc("/join", handler.Join).Methods("POST")
	r.HandleFunc("/heartbeat", handler.Heartbeat).Methods("POST")

	r.HandleFunc("/files/{path}", handler.GetFile).Methods("GET")
	r.HandleFunc("/files", handler.CreateFile).Methods("POST")

	server := http.Server{
		Handler: r,
		Addr:    ":8000",
	}

	log.Println("Starting metadata HTTP server")

	log.Fatal(server.ListenAndServe())
}
