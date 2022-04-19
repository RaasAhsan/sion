package storage

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/RaasAhsan/sion/fs"
	"github.com/gorilla/mux"
)

func StartStorageProcess() {
	client := &http.Client{
		Timeout: 3 * time.Second,
	}
	// TODO: what is the common pattern for this?
	baseUrl := "http://localhost:8000"
	localUrl := "http://localhost:8080"
	nodeId := Join(client, baseUrl, localUrl)
	done := make(chan bool)
	go HeartbeatLoop(client, baseUrl, nodeId, done)
	StartStorageServer()
}

type StorageHandler struct{}

func GatherChunkInventory() {

}

func Join(client *http.Client, baseUrl string, localUrl string) fs.NodeId {
	log.Println("Registering node with master")
	req := fs.RegisterRequest{
		Address: fs.NodeAddress(localUrl),
	}
	reqBytes, err := json.Marshal(req)
	if err != nil {
		log.Fatalln("Failed to serialize register request")
	}
	resp, err := client.Post(fmt.Sprintf("%s/join", baseUrl), "application/json", bytes.NewBuffer(reqBytes))
	if err != nil {
		log.Fatalln("Failed to register")
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		log.Fatalf("Unsuccessful register %d\n", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln("Failed to read register body")
	}
	bodyStr := string(body)
	log.Printf("Successfully registered node: %s", bodyStr)
	return fs.NodeId(bodyStr)
}

// TODO: create an exit channel
func HeartbeatLoop(client *http.Client, baseUrl string, nodeId fs.NodeId, done chan bool) {
	log.Println("Starting heartbeat process")

	ticker := time.NewTicker(5 * time.Second)

	func() {
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				func() {
					req := fs.HeartbeatRequest{NodeId: nodeId}
					reqBody, err := json.Marshal(req)
					if err != nil {
						log.Println("Failed to create request")
						return
					}
					resp, err := client.Post(fmt.Sprintf("%s/heartbeat", baseUrl), "application/json", bytes.NewBuffer(reqBody))
					if err != nil {
						log.Println("Failed to send heartbeat")
						return
					}
					defer resp.Body.Close()
					body, err := io.ReadAll(resp.Body)
					if err != nil {
						log.Println("Failed to read body")
						return
					}
					log.Println(string(body))
				}()
			}
		}
	}()

	ticker.Stop()
}

func StartStorageServer() {
	r := mux.NewRouter()

	h := &StorageHandler{}

	r.HandleFunc("/chunks/{chunkId}", h.DownloadChunk).Methods("GET")
	r.HandleFunc("/chunks/{chunkId}", h.UploadChunk).Methods("POST")

	server := http.Server{
		Handler: r,
		Addr:    ":8080",
	}

	log.Println("Starting storage HTTP server")

	log.Fatal(server.ListenAndServe())
}
