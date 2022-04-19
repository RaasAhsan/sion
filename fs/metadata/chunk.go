package metadata

import (
	"time"

	"github.com/RaasAhsan/sion/fs"
	"github.com/google/uuid"
)

type Chunk struct {
	id          fs.ChunkId
	timeCreated int64
	size        uint
}

func NewChunk() *Chunk {
	return &Chunk{
		id:          fs.ChunkId(uuid.New().String()),
		timeCreated: time.Now().Unix(),
		size:        0,
	}
}
