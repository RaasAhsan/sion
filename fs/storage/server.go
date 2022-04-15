package storage

import (
	"io"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

func StartStorageProcess() {
	client := &http.Client{
		Timeout: 3 * time.Second,
	}
	Join(client)
	go HeartbeatLoop(client)
	server()
}

type StorageHandler struct{}

func Join(client *http.Client) {
	log.Println("Registering node with master")
	resp, err := client.Post("http://localhost:8000/join", "application/json", nil)
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
	log.Printf("Successfully registered node: %s", string(body))
}

// TODO: create an exit channel
func HeartbeatLoop(client *http.Client) {
	log.Println("Starting heartbeat process")
	for {
		func() {
			resp, err := client.Post("http://localhost:8000/heartbeat", "application/json", nil)
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
		time.Sleep(5 * time.Second)
	}
}

func server() {
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
