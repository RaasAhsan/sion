package metadata

// Handles all placement decisioning
// TODO: Do we need some asynchronous process to regulate the cluster?
// Or make decisions when responding to node heartbeats?

// TODO: do we need to distinguish current state of the cluster and the state we want to converge to?
type Placement struct {
	chunkAssignments map[ChunkId]ChunkPlacement
	nodeAssignments  map[NodeId][]ChunkId
}

type ChunkPlacement struct {
	chunkId  ChunkId
	replicas map[NodeId]ReplicaStatus
}

func (p *ChunkPlacement) ReplicaCount() int {
	return len(p.replicas)
}

type ReplicaStatus int

// State machine is Unavailable -> Available
const (
	Unavailable ReplicaStatus = iota
	Available
)

func (p *Placement) NodeJoin(nodeId NodeId) {
	p.nodeAssignments[nodeId] = make([]ChunkId, 0)
	// TODO: allow a chunk to report its chunks on registration
}

func (p *Placement) NodeLeave(nodeId NodeId) {
	for _, chunkId := range p.nodeAssignments[nodeId] {
		placement := p.chunkAssignments[chunkId]
		delete(placement.replicas, nodeId)
	}
}

func AssignChunkToNode(chunkId ChunkId, nodeId NodeId) {

}

func UnassignChunkFromNode(chunkId ChunkId, nodeId NodeId) {

}
