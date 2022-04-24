package cmd

import (
	"log"

	"github.com/RaasAhsan/sion/fs/metadata"
	"github.com/RaasAhsan/sion/fs/storage"
	"github.com/spf13/cobra"
)

var enableStorage bool
var enableMetadata bool

var startCmd = &cobra.Command{
	Use: "start",
	Run: func(cmd *cobra.Command, args []string) {
		if !enableStorage && !enableMetadata {
			panic("At least one of storage or metadata must be specified")
		}

		if enableMetadata {
			ready := make(chan int)
			go metadata.StartMetadataProcess(ready)
			<-ready
			log.Println("Ready to accept connections on metadata server")
		}
		if enableStorage {
			ready := make(chan int)
			go storage.StartStorageProcess(ready)
			<-ready
			log.Println("Ready to accept connections on storage server")
		}

		select {}
	},
}

func init() {
	startCmd.Flags().BoolVarP(&enableStorage, "storage", "s", false, "")
	startCmd.Flags().BoolVarP(&enableMetadata, "metadata", "m", false, "")
	rootCmd.AddCommand(startCmd)
}
