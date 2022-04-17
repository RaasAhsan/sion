package metadata

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func StartMetadataProcess() {
	server()
}

type MetadataHandler struct {
	Namespace *Namespace
	Cluster   *Cluster
}

// TODO: Locate this business logic to another file, and just parse request/render response here?

func (h *MetadataHandler) GetFile(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	path := Path(params["path"])

	// TODO: scope this in a function to minimize critical region?
	h.Namespace.lock.Lock()
	defer h.Namespace.lock.Unlock()

	if !h.Namespace.FileExists(path) {
		http.Error(w, "File does not exist", http.StatusNotFound)
		return
	}

	file := h.Namespace.GetFile(path)

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
	h.Namespace.lock.Lock()
	defer h.Namespace.lock.Unlock()

	if h.Namespace.FileExists(create.Path) {
		http.Error(w, "File already exists", http.StatusBadRequest)
		return
	}

	file := NewFile(create.Path)
	h.Namespace.AddFile(file)

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

	handler := &MetadataHandler{Namespace: NewNamespace(), Cluster: NewCluster()}

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
