package metadata

import (
	"sync"
	"time"
)

type Path string

// TODO: Use a RWMutex to synchronize access to the namespace
type Namespace struct {
	files map[Path]*File
	// TODO: switch to RWMutex
	sync.Mutex
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
	mappings     []*Chunk
}

func (f *File) AppendChunk(c *Chunk) {
	f.mappings = append(f.mappings, c)
}

func NewFile(path Path) *File {
	return &File{
		Path:         path,
		TimeCreated:  time.Now().Unix(),
		TimeModified: time.Now().Unix(),
		Size:         0,
		mappings:     make([]*Chunk, 0),
	}
}
