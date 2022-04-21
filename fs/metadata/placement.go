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
	messages chan PlacementMessage
	sync.Mutex
}

func NewPlacement(messages chan PlacementMessage) *Placement {
	p := &Placement{
		chunkPlacements: make(map[fs.ChunkId]*ChunkPlacement),
		nodeAssignments: make(map[fs.NodeId]*NodeAssignment),
		messages:        messages,
	}
	go p.ProcessMessages()
	return p
}

// TODO: we can use the http.Handler pattern where the message itself
// has a handler method which we call in this method without
// performing a type switch.
func (p *Placement) ProcessMessages() {
	log.Printf("Started placement message processing")
	for {
		msg := <-p.messages
		func() {
			p.Lock()
			defer p.Unlock()
			switch m := msg.(type) {
			case PlacementNodeJoin:
				p.NodeJoin(m.NodeId)
			case PlacementNodeLeave:
				p.NodeLeave(m.NodeId)
			default:
			}
		}()
	}
}

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

// TODO: if we aren't going to have any more message types, we can
// collapse these into one struct probably
type PlacementMessage interface {
	PlacementMessage()
}

type PlacementNodeJoin struct {
	NodeId fs.NodeId
}

func (PlacementNodeJoin) PlacementMessage() {}

type PlacementNodeLeave struct {
	NodeId fs.NodeId
}

func (PlacementNodeLeave) PlacementMessage() {}

// State machine is Unavailable -> Available
const (
	Unavailable ReplicaStatus = iota
	Available
)

func AssignChunkToNode(chunkId fs.ChunkId, nodeId fs.NodeId) {

}

func UnassignChunkFromNode(chunkId fs.ChunkId, nodeId fs.NodeId) {

}
