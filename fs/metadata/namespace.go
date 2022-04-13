package metadata

type ChunkId string
type Path string
type Replica string

type Namespace struct {
	files map[Path]File
}

type File struct {
	path   Path
	chunks []Chunk
	size   uint // TODO: can file size be determined easily?
}

type Chunk struct {
	id       ChunkId
	location Replica
	status   ChunkStatus
	size     uint
}

type ChunkStatus int

const (
	Unavailable ChunkStatus = iota
	Available
)
