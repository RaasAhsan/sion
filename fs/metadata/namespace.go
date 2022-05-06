package metadata

import (
	"errors"
	"sync"
	"time"

	"github.com/RaasAhsan/sion/fs"
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

// Guaranteed to have at least one chunk
func (f *File) Head() *Chunk {
	return f.mappings[0]
}

// Guaranteed to have at least one chunk
func (f *File) Tail() *Chunk {
	return f.mappings[len(f.mappings)-1]
}

func (f *File) FreezeChunk(chunkId fs.ChunkId) bool {
	tail := f.Tail()
	if tail.id != chunkId {
		return false
	}

	tail.Freeze()
	return true
}

func (f *File) AppendChunk(c *Chunk) {
	f.mappings = append(f.mappings, c)
}

func (f *File) FreezeAndAppend(tailChunkId fs.ChunkId) (*Chunk, error) {
	if !f.FreezeChunk(tailChunkId) {
		return nil, errors.New("expected tail chunk")
	}

	c := NewChunk()
	f.AppendChunk(c)
	return c, nil
}

// Create a new file and allocate the first open chunk
func NewFile(path Path) *File {
	head := NewChunk()
	return &File{
		Path:         path,
		TimeCreated:  time.Now().Unix(),
		TimeModified: time.Now().Unix(),
		Size:         0,
		mappings:     []*Chunk{head},
	}
}
