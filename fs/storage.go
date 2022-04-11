package fs

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"

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

func server() {
	r := mux.NewRouter()

	r.HandleFunc("/blocks/{blockId}", func(w http.ResponseWriter, r *http.Request) {
		params := mux.Vars(r)
		blockId := params["blockId"]

		buf := make([]byte, 256)
		bytes := 0

		for {
			n, err := r.Body.Read(buf)
			bytes += n
			if err == io.EOF {
				break
			} else if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}

		log.Printf("Writing block %s, read %d bytes\n", blockId, bytes)
		w.Write([]byte(fmt.Sprintf("%d", bytes)))
	}).Methods("POST")

	server := http.Server{
		Handler: r,
		Addr:    ":8080",
	}

	log.Println("Starting storage HTTP server")

	log.Fatal(server.ListenAndServe())
}
