package storage

import (
	"bufio"
	"fmt"
	"hash/crc32"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/RaasAhsan/sion/fs"
	"github.com/RaasAhsan/sion/fs/api"
	"github.com/gorilla/mux"
)

type NodeState struct {
	Sequence int
	Commands []fs.Command
	Chunks   map[fs.ChunkId]chunk
}

type chunk struct {
	Id fs.ChunkId
}

func (h *StorageHandler) DownloadChunk(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	chunkId := params["chunkId"]

	filename := fmt.Sprintf("./testdir/data/%s", chunkId)
	fi, err := os.Stat(filename)
	if err != nil {
		api.HttpError(w, "Chunk not found", api.ChunkNotFound, http.StatusNotFound)
		return
	}

	// TODO: what if length is too big?
	len := fi.Size()

	// Open chunk file for writing
	in, err := os.Open(filename)
	if err != nil {
		api.HttpError(w, "Chunk not found", api.Unknown, http.StatusInternalServerError)
		return
	}
	defer in.Close()

	w.Header().Add("content-length", fmt.Sprintf("%d", len))
	bytes, err := io.Copy(w, in)
	if err != nil {
		log.Println(err)
		return
	}

	log.Printf("Reading chunk %s, wrote %d bytes\n", chunkId, bytes)
}

func (h *StorageHandler) UploadChunk(w http.ResponseWriter, r *http.Request) {
	if r.ContentLength > fs.ChunkSize {
		api.HttpError(w, "Chunk exceeds max size", api.Unknown, http.StatusBadRequest)
		return
	}

	params := mux.Vars(r)
	chunkId := fs.ChunkId(params["chunkId"])

	// Open chunk file for writing
	filename := fmt.Sprintf("./testdir/data/%s", chunkId)
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
