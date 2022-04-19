package metadata

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/RaasAhsan/sion/fs"
	"github.com/gorilla/mux"
)

func StartMetadataProcess() {
	server()
}

type MetadataHandler struct {
	Namespace *Namespace
	Cluster   *Cluster
	Placement *Placement
}

// TODO: Locate this business logic to another file, and just parse request/render response here?

func (h *MetadataHandler) GetFile(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	path := Path(params["path"])

	// TODO: scope this in a function to minimize critical region?
	h.Namespace.Lock()
	defer h.Namespace.Unlock()

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

// // TODO: just inline this to CreateFile?
// type CreateFile struct {
// 	Path Path
// }
// var create CreateFile
// err := json.NewDecoder(r.Body).Decode(&create)
// if err != nil {
// 	http.Error(w, "Failed to parse body", http.StatusBadRequest)
// 	return
// }

func (h *MetadataHandler) CreateFile(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	path := Path(params["path"])

	// TODO: scope this in a function to minimize critical region?
	h.Namespace.Lock()
	defer h.Namespace.Unlock()

	if h.Namespace.FileExists(path) {
		http.Error(w, "File already exists", http.StatusBadRequest)
		return
	}

	file := NewFile(path)
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

func (h *MetadataHandler) AppendChunkToFile(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	path := Path(params["path"])

	// TODO: scope this in a function to minimize critical region?
	h.Namespace.Lock()
	defer h.Namespace.Unlock()

	if !h.Namespace.FileExists(path) {
		http.Error(w, "File does not exist", http.StatusBadRequest)
		return
	}

	chunk := NewChunk()

	nodeId := func() fs.NodeId {
		h.Placement.Lock()
		defer h.Placement.Unlock()
		return h.Placement.PlaceChunk(chunk)
	}()

	address := func() fs.NodeAddress {
		h.Cluster.Lock()
		defer h.Cluster.Unlock()
		// TODO: handle errors
		return h.Cluster.GetNode(nodeId).Address
	}()

	w.Write([]byte(string(address)))
}

func server() {
	r := mux.NewRouter()

	placementRequests := make(chan fs.NodeId)
	handler := &MetadataHandler{
		Namespace: NewNamespace(),
		Cluster:   NewCluster(placementRequests),
		Placement: NewPlacement(placementRequests),
	}

	r.HandleFunc("/join", handler.Join).Methods("POST")
	r.HandleFunc("/heartbeat", handler.Heartbeat).Methods("POST")

	r.HandleFunc("/files/{path}", handler.GetFile).Methods("GET")
	r.HandleFunc("/files/{path}", handler.CreateFile).Methods("POST")
	r.HandleFunc("/files/{path}/chunks", handler.AppendChunkToFile).Methods("POST")

	server := http.Server{
		Handler: r,
		Addr:    ":8000",
	}

	log.Println("Starting metadata HTTP server")

	log.Fatal(server.ListenAndServe())
}
