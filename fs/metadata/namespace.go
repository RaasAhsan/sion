package metadata

import "github.com/RaasAhsan/sion/fs"

type Path string

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
	id          fs.ChunkId
	timeCreated int64
	size        uint
}

func GetFile() {

}

func NewFile() {

}

func AddFileChunk() {

}

func CommitChunk() {

}
