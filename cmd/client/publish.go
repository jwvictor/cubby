package main

import (
	"fmt"
	"github.com/jwvictor/cubby/pkg/types"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"log"
	url2 "net/url"
	"strings"
)

var (
	publishPublicationId string   = ""
	publishPostId        string   = ""
	publishPermissions   []string = nil
)

var publishCmd = &cobra.Command{
	Use:   "publish",
	Short: "Manage publications and shared blobs",
	Long:  `Publish, manage, and view shared or published blobs.`,
}

var putPublicationCmd = &cobra.Command{
	Use:   "put",
	Short: "Publish a blob from Cubby",
	Long:  `Publish a blob, given its path or blob ID.`,
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := getClient()
		err := client.Authenticate()
		if err != nil {
			log.Printf("Error: Authentication - %s\n", err.Error())
		}
		blobId := args[0]
		b, err := client.GetBlobById(blobId)
		if b == nil {
			fmt.Printf("Could not find blob by ID: %s (%s)\n", blobId, err.Error())
			return
		}

		//if checkIfEncryptedAndEmpty(b) {
		//	fmt.Printf("You are attempting to share a blob that has encrypted body text. Please create an unencrypted blob (-C=none) and share that instead.\n")
		//	return
		//}

		postId := publishPostId
		if postId == "" {
			postId = types.SanitizePostId(b.Title)
		}
		var perms []types.VisibilitySetting
		for _, x := range publishPermissions {
			if strings.ToLower(x) == "public" {
				perms = append(perms, types.VisibilitySetting{
					Type:     types.Public,
					Audience: "",
				})
			} else {
				perms = append(perms, types.VisibilitySetting{
					Type:     types.SingleUser,
					Audience: x,
				})
			}
		}
		if len(perms) == 0 {
			fmt.Printf("No permissions provided - defaulting to public...\n")
			perms = append(perms, types.VisibilitySetting{
				Type: types.Public,
			})
		}
		post := &types.Post{
			Id:            postId,
			OwnerId:       "", // server will fill this out
			Visibility:    perms,
			BlobId:        b.Id,
			PublicationId: publishPublicationId,
		}
		pubPostId, err := client.PutPost(post)
		if err != nil {
			fmt.Printf("Error publishing post: %s\n", err.Error())
			return
		} else {
			fmt.Printf("Published successfully, getting URL...\n")
		}

		userDat, err := client.UserProfile()
		if err != nil {
			fmt.Printf("Error getting display name for URL: %s\n", err.Error())
			return
		}

		host, port := viper.GetString(CfgHost), viper.GetInt(CfgPort)
		url := fmt.Sprintf("%s:%d/v1/post/%s/%s", host, port, url2.QueryEscape(userDat.DisplayName), url2.QueryEscape(pubPostId))
		urlView := fmt.Sprintf("%s:%d/v1/post/%s/%s/view", host, port, url2.QueryEscape(userDat.DisplayName), url2.QueryEscape(pubPostId))
		fmt.Printf("Web: %s\nAPI: %s\n", urlView, url)
	},
}

func checkIfEncryptedAndEmpty(blob *types.Blob) bool {
	if blob.Data == "" {
		for _, x := range blob.RawData {
			if x.Type == types.EncryptedBody {
				return true
			}
		}
	}
	return false
}
