package api

import (
	"github.com/RaasAhsan/sion/fs"
)

type ChunkLocation struct {
	Id    fs.ChunkId
	Nodes []fs.NodeId
}

type FileResponse struct {
	Path         fs.Path
	TimeCreated  int64
	TimeModified int64
	Size         uint
	TailChunk    ChunkLocation
}
