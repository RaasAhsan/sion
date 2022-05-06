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
	files sync.Map
}

func NewNamespace() *Namespace {
	return &Namespace{}
}

func (n *Namespace) FileExists(path Path) bool {
	_, ok := n.files.Load(path)
	return ok
}

// Creates a new file atomically, or returns if it already exists.
func (n *Namespace) CreateFile(newFile *File) bool {
	_, loaded := n.files.LoadOrStore(newFile.Path, newFile)
	return !loaded
}

func (n *Namespace) GetFile(path Path) *File {
	file, ok := n.files.Load(path)
	if !ok {
		return nil
	}

	return file.(*File)
}

// File inode
type File struct {
	Path         Path
	TimeCreated  int64
	TimeModified int64
	Size         uint // TODO: can file size be determined easily?
	mappings     []*Chunk
	sync.RWMutex
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
