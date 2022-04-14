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

// TODO: return content length
func downloadChunk(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	chunkId := params["chunkId"]

	buf := make([]byte, fs.BufferSize)
	bytesRead := 0
	bytesWritten := 0

	// Open chunk file for writing
	filename := fmt.Sprintf("./testdir/data/%s", chunkId)
	in, err := os.Open(filename)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// TODO: pipe approach?
	eof := false
	for !eof {
		n, err := in.Read(buf)
		bytesRead += n
		if err == io.EOF {
			eof = true
		} else if err != nil {
			log.Println(err)
			return
		}

		if !eof {
			m, err := w.Write(buf[:n])
			bytesWritten += m
			if err != nil {
				log.Println(err)
				return
			}
		}
	}

	log.Printf("Reading chunk %s, read %d bytes, wrote %d bytes\n", chunkId, bytesRead, bytesWritten)
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
	// TODO: Do we need to flush?
	filename := fmt.Sprintf("./testdir/data/%s", chunkId)
	out, err := os.Create(filename)
	if err != nil {
		http.Error(w, "Failed to create file", http.StatusInternalServerError)
		return
	}
	defer out.Close()

	// TODO: we could do this but we don't get much type information about the error
	// r.Body = http.MaxBytesReader(w, r.Body, int64(ChunkSize))

	crc := crc32.NewIEEE()

	reader := http.MaxBytesReader(w, r.Body, fs.ChunkSize)
	writer := io.MultiWriter(crc, out)

	wb, err := io.Copy(writer, reader)
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

	log.Printf("Writing chunk %s, wrote %d bytes, checksum: %s\n", chunkId, wb, checksum)
	w.Write([]byte(fmt.Sprintf("%d bytes, %s", wb, checksum)))
}
