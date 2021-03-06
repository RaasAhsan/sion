package storage

import (
	"bufio"
	"fmt"
	"hash/crc32"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"strconv"
	"sync"

	"github.com/RaasAhsan/sion/fs"
	"github.com/RaasAhsan/sion/fs/api"
	"github.com/RaasAhsan/sion/fs/util"
	"github.com/gorilla/mux"
)

type Inventory struct {
	Chunks sync.Map
}

func NewInventory() *Inventory {
	return &Inventory{}
}

func (i *Inventory) GetChunk(id fs.ChunkId) *Chunk {
	v, any := i.Chunks.Load(id)
	if any {
		return v.(*Chunk)
	}
	return nil
}

func (i *Inventory) GetOrPutChunk(c *Chunk) *Chunk {
	rc, _ := i.Chunks.LoadOrStore(c.Id, c)
	return rc.(*Chunk)
}

func (i *Inventory) PutChunk(chunk *Chunk) {
	i.Chunks.Store(chunk.Id, chunk)
}

type Chunk struct {
	Id     fs.ChunkId
	Length uint32
	sync.RWMutex
}

func NewChunk(id fs.ChunkId) *Chunk {
	return &Chunk{
		Id:     id,
		Length: 0,
	}
}

func (c *Chunk) Path(directory string) string {
	return path.Join(directory, string(c.Id))
}

func (h *StorageHandler) DownloadChunk(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	chunkId := fs.ChunkId(params["chunkId"])

	chunk := h.Inventory.GetChunk(chunkId)
	if chunk == nil {
		api.HttpError(w, "Chunk not found", api.ChunkNotFound, http.StatusNotFound)
		return
	}

	chunk.RLocker().Lock()
	defer chunk.RLocker().Unlock()

	range_req := r.Header.Get("Range")
	// TODO: we only support a single range now
	ranges, err := util.ParseRange(range_req, int64(chunk.Length))
	if err != nil {
		api.HttpError(w, "Invalid chunk range", api.Unknown, http.StatusRequestedRangeNotSatisfiable)
		return
	}

	has_range := len(ranges) > 0

	filename := chunk.Path(h.DataDirectory)
	in, err := os.Open(filename)
	if err != nil {
		api.HttpError(w, "Chunk not found", api.Unknown, http.StatusInternalServerError)
		return
	}
	defer in.Close()

	if has_range {
		curr_range := ranges[0]
		w.WriteHeader(http.StatusPartialContent)
		w.Header().Add("content-length", strconv.FormatInt(curr_range.Length, 10))
		w.Header().Add("content-range", curr_range.ContentRange(int64(chunk.Length)))
		_, err = in.Seek(curr_range.Start, io.SeekStart)
		if err != nil {
			log.Println(err)
			return
		}
		reader := io.LimitReader(in, curr_range.Length)
		bytes, err := io.Copy(w, reader)
		if err != nil {
			log.Println(err)
			return
		}

		log.Printf("Reading partial chunk %s, %d len\n", chunkId, bytes)
	} else {
		w.WriteHeader(http.StatusOK)
		w.Header().Add("content-length", strconv.FormatUint(uint64(chunk.Length), 10))
		reader := io.LimitReader(in, int64(chunk.Length))
		bytes, err := io.Copy(w, reader)
		if err != nil {
			log.Println(err)
			return
		}

		log.Printf("Reading full chunk %s, %d len\n", chunkId, bytes)
	}
}

func (h *StorageHandler) UploadChunk(w http.ResponseWriter, r *http.Request) {
	if r.ContentLength > fs.ChunkSize {
		api.HttpError(w, "Chunk exceeds max size", api.Unknown, http.StatusBadRequest)
		return
	}

	params := mux.Vars(r)
	chunkId := fs.ChunkId(params["chunkId"])

	checkChunk := h.Inventory.GetChunk(chunkId)
	if checkChunk != nil {
		api.HttpError(w, "Chunk already exists", api.Unknown, http.StatusBadRequest)
		return
	}

	chunk := NewChunk(chunkId)

	// TODO: locking

	// Open chunk file for writing
	filename := chunk.Path(h.DataDirectory)
	f, err := os.Create(filename)
	if err != nil {
		api.HttpError(w, "Failed to create file", api.Unknown, http.StatusInternalServerError)
		return
	}
	defer f.Close()

	crc := crc32.NewIEEE()

	reader := http.MaxBytesReader(w, r.Body, fs.ChunkSize)
	// Use a buffered writer to minimize number of write syscalls
	// since data may be coming in slowly
	bufferedWriter := bufio.NewWriter(f)
	writer := io.MultiWriter(crc, bufferedWriter)

	bytes, err := io.Copy(writer, reader)
	if err != nil {
		// TODO: OK to delete the file while it is still open?
		log.Printf("Chunk %s encountered an error; deleting...\n", filename)
		err = os.Remove(filename)
		if err != nil {
			log.Printf("Failed to delete chunk %s\n", filename)
		}

		api.HttpError(w, "Failed to copy chunk", api.Unknown, http.StatusInternalServerError)
		return
	}

	// Flush application buffer to OS
	err = bufferedWriter.Flush()
	if err != nil {
		api.HttpError(w, "Failed to flush chunk", api.Unknown, http.StatusInternalServerError)
		return
	}

	// Sync OS buffer to disk
	err = f.Sync()
	if err != nil {
		api.HttpError(w, "Failed to sync chunk", api.Unknown, http.StatusInternalServerError)
		return
	}

	chunk.Length = uint32(bytes)
	// TODO: this is not atomic with respect to chunk creation
	h.Inventory.PutChunk(chunk)

	checksum := fmt.Sprintf("%x", crc.Sum32())

	type response struct {
		Id       fs.ChunkId
		Received int64
		Checksum string
	}

	resp := response{
		Id:       chunkId,
		Received: bytes,
		Checksum: checksum,
	}

	log.Printf("Writing chunk %s, wrote %d bytes, checksum: %s\n", chunkId, bytes, checksum)

	api.HttpOk(w, resp)
}

// AppendChunk implements "optimistic short append".
// Atomically creates a chunk if it doesn't exist to it,
// appends the user payload to it, or if there is not enough space,
// completes it with padding.
func (h *StorageHandler) AppendChunk(w http.ResponseWriter, r *http.Request) {
	payloadLength := r.ContentLength
	if payloadLength == -1 {
		api.HttpError(w, "Content-length must be specified", api.Unknown, http.StatusBadRequest)
		return
	} else if payloadLength > fs.MaxShortAppendLength {
		api.HttpError(w, "Payload length is too large", api.Unknown, http.StatusBadRequest)
		return
	}

	params := mux.Vars(r)
	chunkId := fs.ChunkId(params["chunkId"])

	chunk := h.Inventory.GetOrPutChunk(NewChunk(chunkId))

	chunk.Lock()
	defer chunk.Unlock()

	// TODO: Chunk-level locking, could also increment chunk size and undo if fails

	// First check if chunk is already complete
	if chunk.Length == fs.ChunkSize {
		// TODO: should we return a better status code here, to inform a writer that they need to append?
		api.HttpError(w, "Chunk is full", api.Unknown, http.StatusBadRequest)
		return
	}

	filename := chunk.Path(h.DataDirectory)

	// Then check if we need to pad the chunk and close it
	if chunk.Length+uint32(payloadLength) > fs.ChunkSize {
		err := os.Truncate(filename, fs.ChunkSize)
		if err != nil {
			log.Println(err)
			api.HttpError(w, "Invalid chunk", api.Unknown, http.StatusBadRequest)
		} else {
			chunk.Length = fs.ChunkSize
			api.HttpNoContent(w, "")
		}
		return
	}

	// Otherwise, we can append the contents of the file and update the length
	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		api.HttpError(w, "Failed to get file handle", api.Unknown, http.StatusInternalServerError)
		return
	}
	defer file.Close()

	// Delete any uncommitted data so we can append from the correct position
	err = file.Truncate(int64(chunk.Length))
	if err != nil {
		api.HttpError(w, "Failed to truncate", api.Unknown, http.StatusBadRequest)
		return
	}

	// TODO: deduplicate all this

	reader := http.MaxBytesReader(w, r.Body, fs.ChunkSize)
	// If we can guarantee a single write, no locking needs to happen
	writer := bufio.NewWriter(file)

	bytes, err := io.Copy(writer, reader)
	if err != nil {
		// TODO: OK to delete the file while it is still open?
		log.Printf("Chunk %s encountered an error; deleting...\n", filename)
		// TODO: just undo up to previous length
		err = os.Remove(filename)
		if err != nil {
			log.Printf("Failed to delete chunk %s\n", filename)
		}

		api.HttpError(w, "Failed to copy chunk", api.Unknown, http.StatusInternalServerError)
		return
	}

	// Flush application buffer to OS
	err = writer.Flush()
	if err != nil {
		api.HttpError(w, "Failed to flush chunk", api.Unknown, http.StatusInternalServerError)
		return
	}

	// Sync OS buffer to disk
	err = file.Sync()
	if err != nil {
		api.HttpError(w, "Failed to sync chunk", api.Unknown, http.StatusInternalServerError)
		return
	}

	offset := chunk.Length
	chunk.Length += uint32(bytes)

	// Chunk is now considered committed

	type response struct {
		Offset uint32
		Length uint32
	}

	resp := response{
		Offset: offset,
		Length: uint32(bytes),
	}

	api.HttpOk(w, resp)
}
