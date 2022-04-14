package metadata

// These data structures are used to track the state of the cluster.
// Cluster system can dispatch operations to the namespace (e.g. remove all assignments for a node.)

type NodeId string

type Cluster struct {
	nodes map[NodeId]*node
}

type node struct {
	id                NodeId
	status            nodeStatus
	timeJoined        int64
	timeLastHeartbeat int64
	chunksTotal       uint
	chunksUsed        uint
}

type nodeStatus int

const (
	Online nodeStatus = iota
	Offline
	Decommissioned
)

func RegisterNode() {

}

func HeartbeatNode() {

}
