package cmd

import (
	"context"
	"time"

	"github.com/RaasAhsan/sion/fs/metadata"
	"github.com/RaasAhsan/sion/fs/storage"
	"github.com/spf13/cobra"
	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc"
)

var enableStorage bool
var enableMetadata bool

var startCmd = &cobra.Command{
	Use: "start",
	Run: func(cmd *cobra.Command, args []string) {
		if !enableStorage && !enableMetadata {
			panic("At least one of storage or metadata must be specified")
		}

		client := setupEtcd()
		ctx := context.Background()

		if enableStorage {
			go storage.StartStorageServer(client, ctx)
		}
		if enableMetadata {
			go metadata.StartMetadataServer(client, ctx)
		}

		select {}
	},
}

func setupEtcd() *clientv3.Client {
	etcdCfg := clientv3.Config{
		Endpoints:   []string{"localhost:2379"},
		DialTimeout: time.Second,
		DialOptions: []grpc.DialOption{grpc.WithBlock()},
	}

	client, err := clientv3.New(etcdCfg)
	if err != nil {
		panic(err)
	}

	return client
}

func init() {
	startCmd.Flags().BoolVarP(&enableStorage, "storage", "s", false, "")
	startCmd.Flags().BoolVarP(&enableMetadata, "metadata", "m", false, "")
	rootCmd.AddCommand(startCmd)
}
