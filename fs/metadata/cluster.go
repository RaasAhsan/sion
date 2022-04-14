package metadata

import "github.com/RaasAhsan/sion/fs"

// These data structures are used to track the state of the cluster.
// Cluster system can dispatch operations to the namespace (e.g. remove all assignments for a node.)

type Cluster struct {
	nodes map[fs.NodeId]*node
}

type node struct {
	id                fs.NodeId
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
