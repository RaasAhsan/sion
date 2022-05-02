package storage

import (
	"log"
	"net/http"
	"time"

	"github.com/RaasAhsan/sion/fs/util"
	"github.com/gorilla/mux"
)

func StartStorageProcess(ready chan int) {
	client := &http.Client{
		Timeout: 3 * time.Second,
	}
	// TODO: what is the common pattern for this?
	baseUrl := "http://localhost:8000"
	localUrl := "http://localhost:8080"
	node := Join(client, baseUrl, localUrl)
	done := make(chan bool)
	// TODO: capture this in a struct or something
	go node.HeartbeatLoop(client, baseUrl, done)
	StartStorageServer(node, ready)
}

type StorageHandler struct {
	Inventory     *Inventory
	Node          *Node
	DataDirectory string
}

func StartStorageServer(node *Node, ready chan int) {
	r := mux.NewRouter()

	h := &StorageHandler{
		Inventory:     NewInventory(),
		Node:          node,
		DataDirectory: "./testdir/data",
	}

	r.HandleFunc("/chunks/{chunkId}", h.DownloadChunk).Methods(http.MethodGet)
	r.HandleFunc("/chunks/{chunkId}", h.UploadChunk).Methods(http.MethodPost)
	r.HandleFunc("/chunks/{chunkId}", h.AppendChunk).Methods(http.MethodPatch)

	server := &http.Server{
		Handler: r,
		Addr:    ":8080",
	}

	log.Println("Starting storage HTTP server")

	log.Fatal(util.ListenAndServeNotify(server, ready))
}
