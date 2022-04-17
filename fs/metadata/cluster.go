package metadata

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/RaasAhsan/sion/fs"
	"github.com/google/uuid"
)

// These data structures are used to track the state of the cluster.
// Cluster system can dispatch operations to the namespace (e.g. remove all assignments for a node.)

type Cluster struct {
	Nodes map[fs.NodeId]*Node
}

func NewCluster() *Cluster {
	return &Cluster{
		Nodes: make(map[fs.NodeId]*Node),
	}
}

func (c *Cluster) GetNode(id fs.NodeId) *Node {
	return c.Nodes[id]
}

func (c *Cluster) AddNode(id fs.NodeId) {
	node := &Node{
		Id:                id,
		Status:            Online,
		TimeJoined:        time.Now().Unix(),
		TimeLastHeartbeat: time.Now().Unix(),
		ChunksTotal:       0,
		ChunksUsed:        0,
		HeartbeatChannel:  make(chan bool),
	}

	go node.Monitor()

	c.Nodes[id] = node

	log.Printf("Node %s joined cluster\n", id)
}

func (c *Cluster) HeartbeatNode(id fs.NodeId) error {
	node := c.GetNode(id)
	if node == nil {
		return errors.New("Node does not exist")
	}

	node.Heartbeat()
	return nil
}

type Node struct {
	Id                fs.NodeId
	Status            nodeStatus
	TimeJoined        int64
	TimeLastHeartbeat int64
	ChunksTotal       uint
	ChunksUsed        uint
	HeartbeatChannel  chan bool
}

func (node *Node) Monitor() {
	timer := time.NewTimer(fs.NodeTimeout)
	for {
		select {
		case <-node.HeartbeatChannel:
			node.TimeLastHeartbeat = time.Now().Unix()
			timer.Stop()
			timer = time.NewTimer(fs.NodeTimeout)
		case <-timer.C:
			// TODO: delete node
			log.Printf("Node %s died\n", node.Id)
		}
	}
}

func (node *Node) Heartbeat() {
	node.TimeLastHeartbeat = time.Now().Unix()
	node.HeartbeatChannel <- true
}

type nodeStatus int

const (
	Online nodeStatus = iota
	Offline
	Decommissioned
)

func (h *MetadataHandler) Join(w http.ResponseWriter, r *http.Request) {
	h.lock.Lock()
	defer h.lock.Unlock()

	id := fs.NodeId(uuid.New().String())
	h.cluster.AddNode(id)

	w.Write([]byte(id))
}

func (h *MetadataHandler) Heartbeat(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read body", http.StatusBadRequest)
		return
	}

	var heartbeatReq fs.HeartbeatRequest
	err = json.Unmarshal(body, &heartbeatReq)
	if err != nil {
		http.Error(w, "Failed to parse body", http.StatusBadRequest)
		return
	}

	h.lock.Lock()
	defer h.lock.Unlock()

	err = h.cluster.HeartbeatNode(heartbeatReq.NodeId)
	if err != nil {
		http.Error(w, "Invalid node, please register", http.StatusNotFound)
		return
	}

	w.Write([]byte("heartbeat: ok"))
}
