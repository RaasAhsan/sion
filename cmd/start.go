package cmd

import (
	"fmt"
	"log"
	"net/http"

	"github.com/spf13/cobra"
)

var startCmd = &cobra.Command{
	Use: "start",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("started server")

		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("<h1>Welcome to my web server!</h1>"))
		})

		log.Fatal(http.ListenAndServe(":8080", nil))
	},
}

func init() {
	rootCmd.AddCommand(startCmd)
}
