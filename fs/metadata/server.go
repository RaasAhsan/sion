package metadata

import (
	"fmt"
	"log"
	"net/http"

	"github.com/RaasAhsan/sion/fs"
	"github.com/RaasAhsan/sion/fs/api"
	"github.com/RaasAhsan/sion/fs/util"
	"github.com/gorilla/mux"
)

func StartMetadataProcess(ready chan int) {
	server(ready)
}

type MetadataHandler struct {
	Namespace *Namespace
	Cluster   *Cluster
	Placement *Placement
}

// TODO: Locate this business logic to another file, and just parse request/render response here?

func (h *MetadataHandler) GetFile(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	path := fs.Path(params["path"])

	if !h.Namespace.FileExists(path) {
		api.HttpError(w, "The specified file does not exist.", api.FileNotFound, http.StatusNotFound)
		return
	}

	file := h.Namespace.GetFile(path)
	file.RLock()
	defer file.RUnlock()

	tailChunk := file.Tail()

	nodes := func() []fs.NodeId {
		h.Placement.Lock()
		defer h.Placement.Unlock()
		return h.Placement.GetPlacements(tailChunk.id)
	}()

	resp := api.FileResponse{
		Path:         file.Path,
		TimeCreated:  file.TimeCreated,
		TimeModified: file.TimeModified,
		Size:         file.Size,
		TailChunk: api.ChunkLocation{
			Id:    tailChunk.id,
			Nodes: nodes,
		},
	}

	api.HttpOk(w, resp)
}

// TODO: place these in namespace module?
func (h *MetadataHandler) CreateFile(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	path := fs.Path(params["path"])

	file := NewFile(path)
	created := h.Namespace.CreateFile(file)
	if !created {
		api.HttpError(w, "The specified file exists already.", api.FileNotFound, http.StatusNotFound)
		return
	}

	file.Lock()
	defer file.Unlock()

	tailChunk := file.Tail()

	nodeId := func() fs.NodeId {
		h.Placement.Lock()
		defer h.Placement.Unlock()
		return h.Placement.PlaceChunk(tailChunk.id)
	}()

	resp := api.FileResponse{
		Path:         file.Path,
		TimeCreated:  file.TimeCreated,
		TimeModified: file.TimeModified,
		Size:         file.Size,
		TailChunk: api.ChunkLocation{
			Id:    tailChunk.id,
			Nodes: []fs.NodeId{nodeId},
		},
	}

	api.HttpOk(w, resp)
}

// Freezes the tail chunk of a file and appends a new open chunk
func (h *MetadataHandler) FreezeChunk(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	path := fs.Path(params["path"])
	chunkId := fs.ChunkId(params["chunkId"])

	file := h.Namespace.GetFile(path)
	if file == nil {
		api.HttpError(w, "File does not exist", api.Unknown, http.StatusBadRequest)
		return
	}
	file.Lock()
	defer file.Unlock()

	nextChunk, err := file.FreezeAndAppend(chunkId)
	if err != nil {
		api.HttpError(w, "Expected incorrect tail chunk", api.Unknown, http.StatusConflict)
		return
	}

	nodeId := func() fs.NodeId {
		h.Placement.Lock()
		defer h.Placement.Unlock()
		return h.Placement.PlaceChunk(nextChunk.id)
	}()

	resp := api.ChunkLocation{
		Id:    nextChunk.id,
		Nodes: []fs.NodeId{nodeId},
	}

	api.HttpOk(w, resp)
}

func (h *MetadataHandler) GetChunks(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	path := fs.Path(params["path"])

	file := h.Namespace.GetFile(path)
	if file == nil {
		api.HttpError(w, "File does not exist", api.Unknown, http.StatusBadRequest)
		return
	}
	file.RLock()
	defer file.RUnlock()

	chunks := make([]api.ChunkLocation, 0)

	h.Placement.Lock()
	defer h.Placement.Unlock()

	for _, chunk := range file.mappings {
		// TODO: this can error
		placements := h.Placement.GetPlacements(chunk.id)
		chunks = append(chunks, api.ChunkLocation{Id: chunk.id, Nodes: placements})
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

func server(ready chan int) {
	r := mux.NewRouter()

	pmsgs := make(chan PlacementMessage)
	handler := &MetadataHandler{
		Namespace: NewNamespace(),
		Cluster:   NewCluster(pmsgs),
		Placement: NewPlacement(pmsgs),
	}

	r.HandleFunc("/join", handler.Join).Methods(http.MethodPost)
	r.HandleFunc("/heartbeat", handler.Heartbeat).Methods(http.MethodPost)
	r.HandleFunc("/nodes", handler.GetNodeAddresses).Methods(http.MethodGet)

	r.HandleFunc("/files/{path}", handler.GetFile).Methods(http.MethodGet)
	r.HandleFunc("/files/{path}", handler.CreateFile).Methods(http.MethodPost)
	r.HandleFunc("/files/{path}/chunks", handler.GetChunks).Methods(http.MethodGet)

	r.HandleFunc("/files/{path}/chunks/{chunkId}/freeze", handler.FreezeChunk).Methods(http.MethodPost)

	r.HandleFunc("/version", Version).Methods(http.MethodGet)

	server := &http.Server{
		Handler: r,
		Addr:    ":8000",
	}

	log.Fatal(util.ListenAndServeNotify(server, ready))
}
