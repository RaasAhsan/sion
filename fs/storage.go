package fs

import (
	"context"
	"fmt"
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
	_, err = client.Lease.KeepAlive(ctx, leaseResp.ID)
	if err != nil {
		panic(err)
	}

	// go func() {
	// 	for {
	// 		c := <-ch
	// 		fmt.Println(c)
	// 	}
	// }()

	log.Println("Started lease keep-alive process")
}

func ReadBlock(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	blockId := params["blockId"]

	buf := make([]byte, BufferSize)
	bytesRead := 0
	bytesWritten := 0

	// Open block file for writing
	filename := fmt.Sprintf("./testdir/data/%s", blockId)
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

	log.Printf("Reading block %s, read %d bytes, wrote %d bytes\n", blockId, bytesRead, bytesWritten)
}

func WriteBlock(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	blockId := params["blockId"]

	buf := make([]byte, BufferSize)
	bytesRead := 0
	bytesWritten := 0

	// Open block file for writing
	filename := fmt.Sprintf("./testdir/data/%s", blockId)
	out, err := os.Create(filename)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer out.Close()

	eof := false
	for !eof {
		n, err := r.Body.Read(buf)
		bytesRead += n
		if err == io.EOF {
			eof = true
		} else if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// TODO: is the byte array size limited, or will it write the full 128 bytes?
		// TODO: is this check necessary? keep parity with ReadBlock
		if n > 0 {
			m, err := out.Write(buf)
			bytesWritten += m
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}
	}

	log.Printf("Writing block %s, read %d bytes, wrote %d bytes\n", blockId, bytesRead, bytesWritten)
	w.Write([]byte(fmt.Sprintf("%d, %d", bytesRead, bytesWritten)))
}

func server() {
	r := mux.NewRouter()

	r.HandleFunc("/blocks/{blockId}", ReadBlock).Methods("Get")
	r.HandleFunc("/blocks/{blockId}", WriteBlock).Methods("POST")

	server := http.Server{
		Handler: r,
		Addr:    ":8080",
	}

	log.Println("Starting storage HTTP server")

	log.Fatal(server.ListenAndServe())
}
