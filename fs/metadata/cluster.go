package metadata

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/RaasAhsan/sion/fs"
	"github.com/google/uuid"
)

// These data structures are used to track the state of the cluster.
// Cluster system can dispatch operations to the namespace (e.g. remove all assignments for a node.)

type Cluster struct {
	nodes         map[fs.NodeId]*Node
	placementMsgs chan PlacementMessage
	// TODO: switch to RWMutex
	sync.Mutex
}

func NewCluster(placementMsgs chan PlacementMessage) *Cluster {
	return &Cluster{
		nodes:         make(map[fs.NodeId]*Node),
		placementMsgs: placementMsgs,
	}
}

func (c *Cluster) GetNode(id fs.NodeId) *Node {
	return c.nodes[id]
}

func (c *Cluster) GetAllNodes() []*Node {
	nodes := make([]*Node, len(c.nodes))
	i := 0
	for id := range c.nodes {
		nodes[i] = c.nodes[id]
		i += 1
	}
	return nodes
}

// TODO: separate into NewNode and AddNode
func (c *Cluster) AddNode(id fs.NodeId, address fs.NodeAddress) {
	node := &Node{
		Id:                id,
		Status:            Online,
		TimeJoined:        time.Now().Unix(),
		TimeLastHeartbeat: time.Now().Unix(),
		Address:           address,
		ChunksTotal:       0,
		ChunksUsed:        0,
		HeartbeatChannel:  make(chan bool),
	}

	go node.Monitor(c)

	c.nodes[id] = node
	c.placementMsgs <- PlacementNodeJoin{NodeId: node.Id}

	log.Printf("Node %s (%s) joined cluster\n", id, address)
}

func (c *Cluster) DeleteNode(id fs.NodeId) {
	delete(c.nodes, id)
}

func (c *Cluster) HeartbeatNode(id fs.NodeId) error {
	node := c.GetNode(id)
	if node == nil {
		return errors.New("Node does not exist")
	}

	node.Heartbeat()
	return nil
}

type NodeStatus int

const (
	Online NodeStatus = iota
	Offline
	Decommissioned
)

type Node struct {
	Id                fs.NodeId
	Status            NodeStatus
	TimeJoined        int64
	TimeLastHeartbeat int64
	Address           fs.NodeAddress
	// TODO: this may be a placement concern only
	ChunksTotal      uint
	ChunksUsed       uint
	HeartbeatChannel chan bool
}

func (n *Node) Monitor(c *Cluster) {
	timer := time.NewTimer(fs.NodeTimeout)
	for {
		func() {
			select {
			case <-n.HeartbeatChannel:
				c.Lock()
				defer c.Unlock()
				n.TimeLastHeartbeat = time.Now().Unix()
				timer.Stop()
				timer = time.NewTimer(fs.NodeTimeout)
			case <-timer.C:
				c.Lock()
				defer c.Unlock()
				c.DeleteNode(n.Id)
				// By the time placement subsystem processes this message, cluster API will no longer
				// know of this node; in this case, we can assume the node no longer exists.
				c.placementMsgs <- PlacementNodeLeave{NodeId: n.Id}
				log.Printf("Node %s timed out\n", n.Id)
				return
			}
		}()
	}
}

func (node *Node) Heartbeat() {
	node.TimeLastHeartbeat = time.Now().Unix()
	node.HeartbeatChannel <- true
}

func (h *MetadataHandler) Join(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read body", http.StatusBadRequest)
		return
	}

	var req fs.RegisterRequest
	err = json.Unmarshal(body, &req)
	if err != nil {
		http.Error(w, "Failed to parse body", http.StatusBadRequest)
		return
	}

	h.Cluster.Lock()
	defer h.Cluster.Unlock()

	id := fs.NodeId(uuid.New().String())
	h.Cluster.AddNode(id, req.Address)

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

	h.Cluster.Lock()
	defer h.Cluster.Unlock()

	err = h.Cluster.HeartbeatNode(heartbeatReq.NodeId)
	if err != nil {
		http.Error(w, "Invalid node, please register", http.StatusNotFound)
		return
	}

	w.Write([]byte("heartbeat: ok"))
}

func (h *MetadataHandler) GetNodeAddresses(w http.ResponseWriter, r *http.Request) {
	addresses := make(map[fs.NodeId]fs.NodeAddress)

	type response struct {
		Addresses map[fs.NodeId]fs.NodeAddress
	}

	h.Cluster.Lock()
	defer h.Cluster.Unlock()

	nodes := h.Cluster.GetAllNodes()
	for _, node := range nodes {
		addresses[node.Id] = node.Address
	}

	resp := response{Addresses: addresses}

	jsonBytes, err := json.MarshalIndent(resp, "", "  ")
	if err != nil {
		http.Error(w, "Failed to return response", http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.Write(jsonBytes)
}
