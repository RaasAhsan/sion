package metadata

// Handles all placement decisioning
// TODO: Do we need some asynchronous process to regulate the cluster?
// Or make decisions when responding to node heartbeats?

// TODO: do we need to distinguish current state of the cluster and the state we want to converge to?
type Placement struct {
	chunkPlacements map[ChunkId]*chunkPlacement
	nodeAssignments map[NodeId]*nodeAssignment
}

type chunkPlacement struct {
	chunkId  ChunkId
	replicas map[NodeId]ReplicaStatus
}

func (p *chunkPlacement) ReplicaCount() int {
	return len(p.replicas)
}

type nodeAssignment struct {
	id       NodeId
	chunks   []ChunkId
	sequence int
	log      []int
}

type ReplicaStatus int

// State machine is Unavailable -> Available
const (
	Unavailable ReplicaStatus = iota
	Available
)

func (p *Placement) NodeJoin(nodeId NodeId) {
	node := &nodeAssignment{id: nodeId, chunks: make([]ChunkId, 0), sequence: 0, log: make([]int, 0)}
	p.nodeAssignments[nodeId] = node
	// TODO: allow a chunk to report its chunks on registration
}

func (p *Placement) NodeLeave(nodeId NodeId) {
	for _, chunkId := range p.nodeAssignments[nodeId].chunks {
		placement := p.chunkPlacements[chunkId]
		delete(placement.replicas, nodeId)
	}
}

func AssignChunkToNode(chunkId ChunkId, nodeId NodeId) {

}

func UnassignChunkFromNode(chunkId ChunkId, nodeId NodeId) {

}
