package main

import (
	"github.com/spf13/cobra"
	"log"
)

var catCmd = &cobra.Command{
	Use:   "cat blob-path",
	Short: "Cat a blob from Cubby",
	Long:  `Cats a blob from Cubby to stdout.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := getClient()
		client.CheckVersions()
		err := client.Authenticate()
		if err != nil {
			log.Printf("Error: Authentication - %s\n", err.Error())
			return
		}
		blob, err := client.GetBlobById(args[0])
		if err != nil {
			log.Printf("Error: %s\n", err.Error())
			return
		} else {
			view := "stdout"
			//bo := false
			displayBlob(blob, client, &view, nil)
		}
	},
}
