package fs

type ChunkId string
type NodeId string
type NodeAddress string

// Cluster API
// TODO: should we put this code in the server package?

type RegisterRequest struct {
	Address NodeAddress
}

type RegisterResponse struct {
	NodeId NodeId
}

type HeartbeatRequest struct {
	NodeId   NodeId
	Sequence int
}

type HeartbeatResponse struct {
	Sequence int
	Commands []Command
}

type Command struct {
	ChunkId ChunkId
	Present bool
}

// Namespace API

type GetFile struct {
}

type GetChunk struct {
}
