package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"log"
)

var rmCmd = &cobra.Command{
	Use:   "rm blob-paths...",
	Short: "Delete a blob from Cubby",
	Long:  `Deletes a blob from Cubby by specifying either its ID or path.`,
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			fmt.Printf("Must pass at least one argument: the blob's key, which can be an ID, path, or title substring.")
			return
		}
		client := getClient()
		err := client.Authenticate()
		if err != nil {
			log.Printf("Error: Authentication - %s\n", err.Error())
			return
		}
		for _, arg := range args {
			blob, err := client.GetBlobById(arg)
			if err != nil {
				fmt.Printf("Failed to get blob: %s\n", err.Error())
				return
			}
			_, err = client.DeleteBlob(blob.Id)
			if err != nil {
				fmt.Printf("Failed to delete blob: %s\n", err.Error())
				return
			} else {
				fmt.Printf("Deleted blob `%s` with ID %s.\n", blob.Title, blob.Id)
			}
		}
	},
}
