package metadata

import (
	"log"
	"math/rand"
	"sync"

	"github.com/RaasAhsan/sion/fs"
)

// Handles all placement decisioning
// TODO: Do we need some asynchronous process to regulate the cluster?
// Or make decisions when responding to node heartbeats?

// TODO: do we need to distinguish current state of the cluster and the state we want to converge to?
type Placement struct {
	chunkPlacements map[fs.ChunkId]*ChunkPlacement
	nodeAssignments map[fs.NodeId]*NodeAssignment
	// TODO: refine this later
	requests chan fs.NodeId
	lock     sync.Mutex
}

func NewPlacement(requests chan fs.NodeId) *Placement {
	p := &Placement{
		chunkPlacements: make(map[fs.ChunkId]*ChunkPlacement),
		nodeAssignments: make(map[fs.NodeId]*NodeAssignment),
		requests:        requests,
	}
	go p.ProcessRequests()
	return p
}

func (p *Placement) ProcessRequests() {
	log.Printf("Started placement request processing")
	for {
		req := <-p.requests
		func() {
			p.lock.Lock()
			defer p.lock.Unlock()
			p.NodeJoin(req)
		}()
	}
}

func (p *Placement) PlaceChunk(c *Chunk) fs.NodeId {
	// TODO: go 1.19 we can use a generic method
	keys := make([]fs.NodeId, len(p.nodeAssignments))
	i := 0
	for k := range p.nodeAssignments {
		keys[i] = k
	}

	// TODO: create an interface for placement strategy
	node := p.nodeAssignments[keys[rand.Intn(len(keys))]]
	node.Chunks = append(node.Chunks, c.id)

	replicas := make(map[fs.NodeId]ReplicaStatus)
	replicas[node.Id] = Unavailable
	placement := &ChunkPlacement{
		chunkId:  c.id,
		replicas: replicas,
		chunk:    c,
	}
	p.chunkPlacements[c.id] = placement

	return node.Id
}

type ChunkPlacement struct {
	chunkId  fs.ChunkId
	replicas map[fs.NodeId]ReplicaStatus
	chunk    *Chunk
}

func (p *ChunkPlacement) ReplicaCount() int {
	return len(p.replicas)
}

// TODO: could potentially store a pointer to *Node here, better if we consult cluster first though?
type NodeAssignment struct {
	Id       fs.NodeId
	Chunks   []fs.ChunkId
	Sequence int
	Log      []int
}

type ReplicaStatus int

// State machine is Unavailable -> Available
const (
	Unavailable ReplicaStatus = iota
	Available
)

func (p *Placement) NodeJoin(nodeId fs.NodeId) {
	node := &NodeAssignment{
		Id:       nodeId,
		Chunks:   make([]fs.ChunkId, 0),
		Sequence: 0,
		Log:      make([]int, 0),
	}
	p.nodeAssignments[nodeId] = node
	// TODO: allow a chunk to report its chunks on registration
}

func (p *Placement) NodeLeave(nodeId fs.NodeId) {
	for _, chunkId := range p.nodeAssignments[nodeId].Chunks {
		placement := p.chunkPlacements[chunkId]
		delete(placement.replicas, nodeId)
	}
}

func AssignChunkToNode(chunkId fs.ChunkId, nodeId fs.NodeId) {

}

func UnassignChunkFromNode(chunkId fs.ChunkId, nodeId fs.NodeId) {

}
