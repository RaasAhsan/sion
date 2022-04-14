package metadata

type ChunkId string
type Path string
type Replica string

// TODO: Use a RWMutex to synchronize access to the namespace
type Namespace struct {
	files map[Path]File
}

// File inode
type File struct {
	path         Path
	timeCreated  int64
	timeModified int64
	size         uint // TODO: can file size be determined easily?
	chunks       []Chunk
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

func GetFile() {

}

func NewFile() {

}

func AddFileChunk() {

}

func CommitChunk() {

}
