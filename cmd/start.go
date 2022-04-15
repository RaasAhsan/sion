package cmd

import (
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

		if enableStorage {
			go storage.StartStorageProcess()
		}
		if enableMetadata {
			go metadata.StartMetadataProcess()
		}

		select {}
	},
}

func init() {
	startCmd.Flags().BoolVarP(&enableStorage, "storage", "s", false, "")
	startCmd.Flags().BoolVarP(&enableMetadata, "metadata", "m", false, "")
	rootCmd.AddCommand(startCmd)
}
