package fs

import (
	"context"
	"fmt"
	"hash/crc32"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	clientv3 "go.etcd.io/etcd/client/v3"
)

func StartStorage(client *clientv3.Client, ctx context.Context) {
	register(client, ctx, "0")
	server()
}

// Registers the node in etcd and begins a lease keep-alive process
func register(client *clientv3.Client, ctx context.Context, nodeId string) {
	// Grant a lease associated with this node's lifetime
	leaseResp, err := client.Lease.Grant(ctx, 60)
	if err != nil {
		panic(err)
	}

	kvc := clientv3.NewKV(client)

	_, err = kvc.Put(ctx, "/sion/nodes/"+nodeId, "1", clientv3.WithLease(leaseResp.ID))
	if err != nil {
		panic(err)
	}

	log.Printf("Registered node %s in etcd\n", nodeId)

	// TODO: consume keep-alives
	ch, err := client.Lease.KeepAlive(ctx, leaseResp.ID)
	if err != nil {
		panic(err)
	}

	go func() {
		for {
			<-ch
			// c := <-ch
			// fmt.Println(c)
		}
	}()

	log.Println("Started lease keep-alive process")
}

// TODO: return content length
func downloadChunk(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	chunkId := params["chunkId"]

	buf := make([]byte, BufferSize)
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
			m, err := w.Write(buf)
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
	params := mux.Vars(r)
	chunkId := params["chunkId"]

	// Open chunk file for writing
	// TODO: Do we need to flush?
	filename := fmt.Sprintf("./testdir/data/%s", chunkId)
	out, err := os.Create(filename)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer out.Close()

	// TODO: we could do this but we don't get much type information about the error
	// r.Body = http.MaxBytesReader(w, r.Body, int64(ChunkSize))

	crc := crc32.NewIEEE()

	buf := make([]byte, BufferSize)
	rb := 0
	wb := 0

	eof := false
	for !eof {
		n, err := r.Body.Read(buf)
		rb += n
		if err == io.EOF {
			eof = true
		} else if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// TODO: we should infer this from content-length if it is available
		if rb > ChunkSize {
			// TODO: OK to delete the file while it is still open?
			log.Printf("Chunk %s is too large; deleting...\n", filename)
			err = os.Remove(filename)
			if err != nil {
				log.Printf("Failed to delete chunk %s\n", filename)
			}
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Chunk is too large"))
			return
		}

		// TODO: is the byte array size limited, or will it write the full 128 bytes?
		// TODO: is this check necessary? keep parity with readChunk
		if n > 0 {
			_, err := crc.Write(buf[0:n])
			if err != nil {
				log.Println("Failed to run crc32 on chunk buffer")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			m, err := out.Write(buf)
			wb += m
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}
	}

	checksum := fmt.Sprintf("%x", crc.Sum32())

	log.Printf("Writing chunk %s, read %d bytes, wrote %d bytes, checksum: %s\n", chunkId, rb, wb, checksum)
	w.Write([]byte(fmt.Sprintf("%d, %d, %s", rb, wb, checksum)))
}

func server() {
	r := mux.NewRouter()

	r.HandleFunc("/chunks/{chunkId}", downloadChunk).Methods("GET")
	r.HandleFunc("/chunks/{chunkId}", uploadChunk).Methods("POST")

	server := http.Server{
		Handler: r,
		Addr:    ":8080",
	}

	log.Println("Starting storage HTTP server")

	log.Fatal(server.ListenAndServe())
}
