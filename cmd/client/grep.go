package main

import (
	"fmt"
	"github.com/jwvictor/cubby/pkg/types"
	"github.com/spf13/cobra"
	"log"
	"strings"
)

var (
	grepCaseInsensitive = false
)

var grepCmd = &cobra.Command{
	Use:   "grep",
	Short: "Search for a substring",
	Long:  `Search for a substring in blob bodies, including encrypted ones.`,
	Run: func(cmd *cobra.Command, args []string) {
		client := getClient()
		client.CheckVersions()
		err := client.Authenticate()
		if err != nil {
			log.Printf("Error: Authentication - %s\n", err.Error())
		}
		blob, err := client.ListBlobs(nil)
		if err != nil {
			log.Printf("Error: %s\n", err.Error())
			return
		}

		queryStr := args[0]
		var ids []string
		for _, x := range blob {
			ids = append(ids, getAllBlobIds(x)...)
		}
		fmt.Printf("Searching %d blobs...\n", len(ids))
		for _, id := range ids {
			thisBlob, err := client.GetBlobById(id)
			if err != nil {
				fmt.Printf("Failed to get blob: %s\n", err.Error())
				return
			}
			body := thisBlob.Data
			if encData := getBlobEncryptedBody(thisBlob); encData != nil {
				bs, err := decryptData(encData.Data)
				if err != nil {
					fmt.Printf("Failed to decrypt data: %s\n", err.Error())
					return
				} else {
					body = string(bs)
				}
			}

			origBody := body
			if grepCaseInsensitive {
				// Make all comparisons insensitive
				body = strings.ToLower(body)
				queryStr = strings.ToLower(queryStr)
			}

			if strings.Contains(body, queryStr) {
				occIdx := strings.Index(body, queryStr)
				n := 32
				sIdx := occIdx - n
				eIdx := occIdx + n
				if sIdx < 0 {
					sIdx = 0
				}
				if eIdx > len(body) {
					eIdx = len(body)
				}
				preview := origBody[sIdx:eIdx]
				preview = strings.ReplaceAll(preview, "\n", " ")
				fmt.Printf("[%s] %s\n", id, preview)
			}
		}
	},
}

func getAllBlobIds(blob *types.BlobSkeleton) []string {
	out := []string{blob.Id}
	for _, child := range blob.Children {
		out = append(out, getAllBlobIds(child)...)
	}
	return out
}
