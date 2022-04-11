package cmd

import (
	"context"
	"time"

	"github.com/spf13/cobra"
	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc"

	"github.com/RaasAhsan/sion/fs"
)

var storage bool
var metadata bool

var startCmd = &cobra.Command{
	Use: "start",
	Run: func(cmd *cobra.Command, args []string) {
		if !storage && !metadata {
			panic("At least one of storage or metadata must be specified")
		}

		client := setupEtcd()
		ctx := context.Background()

		if storage {
			go fs.StartStorage(client, ctx)
		}
		if metadata {
			go fs.StartMetadata(client, ctx)
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
	startCmd.Flags().BoolVarP(&storage, "storage", "s", false, "")
	startCmd.Flags().BoolVarP(&metadata, "metadata", "m", false, "")
	rootCmd.AddCommand(startCmd)
}
