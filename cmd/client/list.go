package main

import (
	"encoding/json"
	"fmt"
	"github.com/jwvictor/cubby/pkg/types"
	"github.com/spf13/cobra"
	"log"
)

var (
	listCmdRenderJson = false
)

func renderSkeleton(blob *types.BlobSkeleton, indent int) string {
	sep := ". "
	var out string
	for i := 0; i < indent; i++ {
		out += sep
	}
	out += blob.Title + "\n"
	for _, child := range blob.Children {
		out += renderSkeleton(child, indent+1)
	}
	return out
}

func renderJson(blob *types.BlobSkeleton) string {
	bs, err := json.Marshal(blob)
	if err != nil {
		return ""
	}
	return string(bs)
}

func mergeBlobs(blobs []*types.BlobSkeleton) *types.BlobSkeleton {
	return &types.BlobSkeleton{
		Id:       "root",
		Title:    "root",
		Tags:     []string{"system"},
		Children: blobs,
		OwnerId:  "",
	}
}

var listCmd = &cobra.Command{
	Use:   "list [sub-blob]",
	Short: "List a substructure or all blobs in Cubby",
	Long:  `Lists all blobs or blobs under a specific path.`,
	Run: func(cmd *cobra.Command, args []string) {
		client := getClient()
		client.CheckVersions()
		err := client.Authenticate()
		if err != nil {
			log.Printf("Error: Authentication - %s\n", err.Error())
		}
		var listBlobId *string
		if len(args) > 0 {
			blob, err := client.GetBlobById(args[0])
			if err != nil {
				fmt.Printf("Error querying blob: %s\n", args[0])
				return
			}
			listBlobId = &blob.Id
		}
		blob, err := client.ListBlobs(listBlobId)
		if err != nil {
			log.Printf("Error: %s\n", err.Error())
		} else {
			if !listCmdRenderJson {
				for _, x := range blob {
					fmt.Printf(renderSkeleton(x, 1))
				}
			} else {
				fmt.Printf(renderJson(mergeBlobs(blob)))
			}
		}
	},
}
