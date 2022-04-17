package metadata

import (
	"sync"
	"time"

	"github.com/RaasAhsan/sion/fs"
)

type Path string

// TODO: Use a RWMutex to synchronize access to the namespace
type Namespace struct {
	files map[Path]*File
	// TODO: switch to RWMutex
	lock sync.Mutex
}

func NewNamespace() *Namespace {
	return &Namespace{files: make(map[Path]*File)}
}

func (n *Namespace) FileExists(path Path) bool {
	_, exist := n.files[path]
	return exist
}

func (n *Namespace) AddFile(file *File) {
	n.files[file.Path] = file
}

func (n *Namespace) GetFile(path Path) *File {
	return n.files[path]
}

// File inode
type File struct {
	Path         Path
	TimeCreated  int64
	TimeModified int64
	Size         uint // TODO: can file size be determined easily?
	Chunks       []Chunk
}

func NewFile(path Path) *File {
	return &File{
		Path:         path,
		TimeCreated:  time.Now().Unix(),
		TimeModified: time.Now().Unix(),
		Size:         0,
		Chunks:       make([]Chunk, 0),
	}
}

type Chunk struct {
	id          fs.ChunkId
	timeCreated int64
	size        uint
}

func GetFile() {

}

func AddFileChunk() {

}

func CommitChunk() {

}
