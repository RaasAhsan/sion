package api

import "github.com/RaasAhsan/sion/fs"

type ChunkLocation struct {
	Id    fs.ChunkId
	Nodes []fs.NodeId
}
