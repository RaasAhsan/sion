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
	"github.com/RaasAhsan/sion/fs/api"
)

func Join(client *http.Client, baseUrl string, localUrl string) fs.NodeId {
	log.Println("Registering node with master")
	req := api.RegisterRequest{
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
	success, err := api.ParseResponse[api.RegisterResponse](resp.Body)
	if err != nil {
		log.Fatalf("Failed to register")
	}
	log.Printf("Successfully registered node: %s", success.NodeId)
	return success.NodeId
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
					req := api.HeartbeatRequest{NodeId: nodeId}
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
					bytes, err := io.ReadAll(resp.Body)
					if err != nil {
						log.Println("Failed to read body")
						return
					}
					var body api.HeartbeatResponse
					err = json.Unmarshal(bytes, &body)
					if err != nil {
						log.Fatalln("Failed to parse heartbeat body")
					}
					log.Println("heartbeat ok")
				}()
			}
		}
	}()

	ticker.Stop()
}
