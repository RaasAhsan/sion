package fs

import (
	"context"
	"log"
	"net/http"

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
	server := http.Server{
		Addr: ":8080",
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("12345"))
	})

	log.Println("Starting storage HTTP server")

	log.Fatal(server.ListenAndServe())
}
