package cmd

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/spf13/cobra"
	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc"
)

var startCmd = &cobra.Command{
	Use: "start",
	Run: func(cmd *cobra.Command, args []string) {
		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
			w.Write([]byte("12345"))
		})

		etcdCfg := clientv3.Config{
			Endpoints:   []string{"localhost:2379"},
			DialTimeout: time.Second,
			DialOptions: []grpc.DialOption{grpc.WithBlock()},
		}

		client, err := clientv3.New(etcdCfg)
		if err != nil {
			panic(err)
		}

		ctx := context.Background()

		register(client, ctx, "0")

		log.Println("Starting HTTP server")
		server := http.Server{
			Addr: ":8080",
		}

		log.Fatal(server.ListenAndServe())
	},
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
			c := <-ch
			fmt.Println(c)
		}
	}()

	log.Println("Started lease keep-alive process")
}

func init() {
	rootCmd.AddCommand(startCmd)
}
