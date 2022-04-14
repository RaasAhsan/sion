package metadata

// These data structures are used to track the state of the cluster.
// Cluster system can dispatch operations to the namespace (e.g. remove all assignments for a node.)

type NodeId string

type Cluster struct {
	nodes map[NodeId]Node
}

type Node struct {
	id              NodeId
	status          NodeStatus
	timeJoined      int64
	timeLastMessage int64
	chunksTotal     uint
	chunksUsed      uint
}

type NodeStatus int

const (
	Online NodeStatus = iota
	Offline
	Decommissioned
)

func RegisterNode() {

}

func HeartbeatNode() {

}
