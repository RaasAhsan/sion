package storage

import (
	"fmt"
	"hash/crc32"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/RaasAhsan/sion/fs"
	"github.com/gorilla/mux"
)

type chunkId string

type metadata struct {
	chunks map[chunkId]chunk
}

type chunk struct {
	id chunkId
}

func downloadChunk(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	chunkId := params["chunkId"]

	filename := fmt.Sprintf("./testdir/data/%s", chunkId)
	fi, err := os.Stat(filename)
	if err != nil {
		http.Error(w, "Chunk not found", http.StatusNotFound)
		return
	}
	// TODO: what if length is too big?
	len := fi.Size()

	// Open chunk file for writing
	in, err := os.Open(filename)
	if err != nil {
		http.Error(w, "Failed to open chunk", http.StatusInternalServerError)
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

// TODO: assert length
// TODO: split out logic
func uploadChunk(w http.ResponseWriter, r *http.Request) {
	if r.ContentLength > fs.ChunkSize {
		http.Error(w, "Chunk exceeds max size", http.StatusBadRequest)
		return
	}

	params := mux.Vars(r)
	chunkId := params["chunkId"]

	// Open chunk file for writing
	filename := fmt.Sprintf("./testdir/data/%s", chunkId)
	out, err := os.Create(filename)
	if err != nil {
		http.Error(w, "Failed to create file", http.StatusInternalServerError)
		return
	}
	defer out.Close()

	crc := crc32.NewIEEE()

	reader := http.MaxBytesReader(w, r.Body, fs.ChunkSize)
	writer := io.MultiWriter(crc, out)

	bytes, err := io.Copy(writer, reader)
	if err != nil {
		// TODO: OK to delete the file while it is still open?
		log.Printf("Chunk %s encountered an error; deleting...\n", filename)
		err = os.Remove(filename)
		if err != nil {
			log.Printf("Failed to delete chunk %s\n", filename)
		}

		http.Error(w, "Failed to write chunk", http.StatusInternalServerError)
		return
	}

	checksum := fmt.Sprintf("%x", crc.Sum32())

	log.Printf("Writing chunk %s, wrote %d bytes, checksum: %s\n", chunkId, bytes, checksum)
	w.Write([]byte(fmt.Sprintf("%d bytes, %s", bytes, checksum)))
}
