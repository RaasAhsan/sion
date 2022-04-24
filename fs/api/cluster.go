package api

import "github.com/RaasAhsan/sion/fs"

type RegisterRequest struct {
	Address fs.NodeAddress
}

type RegisterResponse struct {
	NodeId fs.NodeId
}

type HeartbeatRequest struct {
	NodeId   fs.NodeId
	Sequence int
}

type HeartbeatResponse struct {
	Sequence int
	Commands []Command
}

type Command struct {
	ChunkId fs.ChunkId
	Present bool
}
