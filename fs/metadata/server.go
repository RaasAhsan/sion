package metadata

import (
	"fmt"
	"log"
	"net/http"

	"github.com/RaasAhsan/sion/fs"
	"github.com/RaasAhsan/sion/fs/api"
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
		api.HttpError(w, "The specified file does not exist.", api.FileNotFound, http.StatusNotFound)
		return
	}

	file := h.Namespace.GetFile(path)

	api.HttpOk(w, file)
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

// TODO: place these in namespace module?
func (h *MetadataHandler) CreateFile(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	path := Path(params["path"])

	// TODO: scope this in a function to minimize critical region?
	h.Namespace.Lock()
	defer h.Namespace.Unlock()

	if h.Namespace.FileExists(path) {
		api.HttpError(w, "The specified file does not exist.", api.FileNotFound, http.StatusNotFound)
		return
	}

	file := NewFile(path)
	h.Namespace.AddFile(file)

	api.HttpOk(w, file)
}

func (h *MetadataHandler) AppendChunk(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	path := Path(params["path"])

	// TODO: scope this in a function to minimize critical region?
	h.Namespace.Lock()
	defer h.Namespace.Unlock()

	file := h.Namespace.GetFile(path)
	if file == nil {
		http.Error(w, "File does not exist", http.StatusBadRequest)
		return
	}

	chunk := NewChunk()
	file.AppendChunk(chunk)

	nodeId := func() fs.NodeId {
		h.Placement.Lock()
		defer h.Placement.Unlock()
		return h.Placement.PlaceChunk(chunk.id)
	}()

	address := func() fs.NodeAddress {
		h.Cluster.Lock()
		defer h.Cluster.Unlock()
		// TODO: handle errors
		return h.Cluster.GetNode(nodeId).Address
	}()

	w.Write([]byte(string(address)))
}

func (h *MetadataHandler) GetChunks(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	path := Path(params["path"])

	// TODO: fix this locking, don't need it while querying placement
	h.Namespace.Lock()
	defer h.Namespace.Unlock()

	file := h.Namespace.GetFile(path)
	if file == nil {
		http.Error(w, "File does not exist", http.StatusBadRequest)
		return
	}

	type chunkLocation struct {
		Id    fs.ChunkId
		Nodes []fs.NodeId
	}

	chunks := make([]chunkLocation, 0)

	h.Placement.Lock()
	defer h.Placement.Unlock()

	for _, chunk := range file.mappings {
		// TODO: this can error
		placements := h.Placement.GetPlacements(chunk.id)
		chunks = append(chunks, chunkLocation{Id: chunk.id, Nodes: placements})
	}

	api.HttpOk(w, chunks)
}

func Version(w http.ResponseWriter, r *http.Request) {
	type version struct {
		ServerVersion string
		MajorVersion  int
		MinorVersion  int
		PatchVersion  int
	}

	majorVersion := 1
	minorVersion := 0
	patchVersion := 0
	serverVersion := fmt.Sprintf("%d.%d.%d", majorVersion, minorVersion, patchVersion)

	body := version{
		ServerVersion: serverVersion,
		MajorVersion:  majorVersion,
		MinorVersion:  minorVersion,
		PatchVersion:  patchVersion,
	}

	api.HttpOk(w, body)
}

func server() {
	r := mux.NewRouter()

	pmsgs := make(chan PlacementMessage)
	handler := &MetadataHandler{
		Namespace: NewNamespace(),
		Cluster:   NewCluster(pmsgs),
		Placement: NewPlacement(pmsgs),
	}

	r.HandleFunc("/join", handler.Join).Methods("POST")
	r.HandleFunc("/heartbeat", handler.Heartbeat).Methods("POST")
	r.HandleFunc("/nodes", handler.GetNodeAddresses).Methods("GET")

	r.HandleFunc("/files/{path}", handler.GetFile).Methods("GET")
	r.HandleFunc("/files/{path}", handler.CreateFile).Methods("POST")
	r.HandleFunc("/files/{path}/chunks", handler.GetChunks).Methods("GET")
	r.HandleFunc("/files/{path}/chunks", handler.AppendChunk).Methods("POST")

	r.HandleFunc("/version", Version).Methods("GET")

	server := http.Server{
		Handler: r,
		Addr:    ":8000",
	}

	log.Println("Starting metadata HTTP server")

	log.Fatal(server.ListenAndServe())
}
