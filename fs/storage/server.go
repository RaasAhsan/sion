package storage

import (
	"io"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

func StartStorageServer() {
	go HeartbeatLoop()
	server()
}

// TODO: create an exit channel
func HeartbeatLoop() {
	log.Println("Starting heartbeat process")
	client := http.Client{
		Timeout: 3 * time.Second,
	}
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

	r.HandleFunc("/chunks/{chunkId}", downloadChunk).Methods("GET")
	r.HandleFunc("/chunks/{chunkId}", uploadChunk).Methods("POST")

	server := http.Server{
		Handler: r,
		Addr:    ":8080",
	}

	log.Println("Starting storage HTTP server")

	log.Fatal(server.ListenAndServe())
}
