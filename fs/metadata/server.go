package metadata

import (
	"context"
	"fmt"
	"log"

	clientv3 "go.etcd.io/etcd/client/v3"
)

func StartMetadataServer(client *clientv3.Client, ctx context.Context) {
	log.Println("Starting metadata server")
	revision := getNodes(client, ctx)

	log.Printf("etcd current revision is %d\n", revision)

	watchChan := client.Watch(ctx, "/sion/nodes/", clientv3.WithRev(revision+1), clientv3.WithPrefix())
	for watchResp := range watchChan {
		for _, ev := range watchResp.Events {
			fmt.Println(ev)
		}
	}
}

func getNodes(client *clientv3.Client, ctx context.Context) int64 {
	resp, err := client.Get(ctx, "/sion/nodes/", clientv3.WithPrefix())
	if err != nil {
		panic(err)
	}

	return resp.Header.Revision
}
