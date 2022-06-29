package main

import (
	"fmt"
	"github.com/jwvictor/cubby/cmd/client/tuiviewer"
	"github.com/jwvictor/cubby/pkg/client"
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
	publishOwnerId       string   = ""
	publishPermissions   []string = nil
)

var publishCmd = &cobra.Command{
	Use:   "publish",
	Short: "Manage publications and shared blobs",
	Long:  `Publish, manage, and view shared or published blobs.`,
}

var getPublicationCmd = &cobra.Command{
	Use:   "get",
	Short: "Gets a post",
	Long:  `Gets a posted blob, given the post ID (title) or blob ID.`,
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := getClient()
		err := client.Authenticate()
		if err != nil {
			log.Printf("Error: Authentication - %s\n", err.Error())
			return
		}
		postId := args[0]
		ownerId := publishOwnerId
		if ownerId == "" {
			uid, err := client.UserProfile()
			if err != nil {
				fmt.Printf("Could not get user profile:%s\n", err.Error())
				return
			}
			ownerId = uid.DisplayName
		}
		post, err := client.GetPostById(ownerId, postId)
		if err != nil {
			fmt.Printf("Could not find post by ID: %s (%s)\n", postId, err.Error())
			return
		}
		cTyp := types.ResolveContentType(post.Blobs[0].Type)
		fileext := ""
		if cTyp != nil {
			fileext = cTyp.FileExtension
		}
		displayPost(post.Body, post.EncryptedBody, post.Posts[0].Id, fileext)
	},
}

func displayPost(body string, encBody []byte, title, fileExt string) {
	if encBody != nil {
		key := viper.GetString(CfgSymmetricKey)
		keyBytes, err := types.DeriveSymmetricKey(key)
		if err != nil {
			fmt.Printf("Error deriving key: %s\n", err.Error())
		}
		plaintxt, err := types.DecryptSymmetric(encBody, keyBytes)
		if err != nil {
			fmt.Printf("Error decrypting: %s\n", err.Error())
		}
		body += string(plaintxt)
	}

	userViewer := viper.GetString(CfgViewer)
	switch userViewer {
	case CfgViewerTui:
		tuiviewer.RunViewer(body, title)
	case CfgViewerEditor:
		_, err := openInEditor(body, title, fileExt)
		if err != nil {
			fmt.Errorf("Failed to open editor: %s\n", err.Error())
		}
	default:
		fmt.Printf(body)
	}
}

var listPublicationsCmd = &cobra.Command{
	Use:   "list",
	Short: "Lists all published blobs",
	Long:  `Lists all currently active published blobs.`,
	Args:  cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		client := getClient()
		err := client.Authenticate()
		if err != nil {
			log.Printf("Error: Authentication - %s\n", err.Error())
			return
		}
		posts, err := client.ListPublishedBlobs()
		if err != nil {
			log.Printf("Error getting published blobs: %s\n", err.Error())
			return
		}

		renderPosts(posts)

	},
}

func renderPermissions(post *types.Post) string {
	var ps []string
	for _, x := range post.Visibility {
		if x.Type == types.Public {
			ps = append(ps, string(types.Public))
		} else {
			ps = append(ps, x.Audience)
		}
	}
	return strings.Join(ps, " ")
}

func renderPosts(posts []*types.Post) {
	for _, post := range posts {
		fmt.Printf("%s - %s [%s] \n", post.Id, post.BlobId, renderPermissions(post))
	}
}

var rmPublicationCmd = &cobra.Command{
	Use:   "rm",
	Short: "Delete a post",
	Long:  `Deletes a posted blob, given the post ID (title) or blob ID.`,
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := getClient()
		err := client.Authenticate()
		if err != nil {
			log.Printf("Error: Authentication - %s\n", err.Error())
			return
		}
		for _, postId := range args {
			uid, err := client.UserProfile()
			if err != nil {
				fmt.Printf("Could not get user profile:%s\n", err.Error())
				return
			}
			post, err := client.GetPostById(uid.DisplayName, postId)
			if err != nil {
				fmt.Printf("Could not find post by ID: %s (%s)\n", postId, err.Error())
				return
			}
			_, err = client.DeletePost(uid.DisplayName, post.Posts[0].Id)
			if err != nil {
				fmt.Printf("Could not delete post by ID: %s (%s)\n", postId, err.Error())
				return
			} else {
				fmt.Printf("Successfully deleted post with ID: %s\n", postId)
			}
		}
	},
}

func resolveUser(userId string, client *client.CubbyClient) (string, error) {
	userRes, err := client.SearchUser(userId)
	if err != nil {
		return "", err
	} else {
		return userRes.Id, nil
	}
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
			return
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
			// TODO: check if there's a name collision and, if so, add some extra characters to the ID
		}
		var perms []types.VisibilitySetting
		for _, x := range publishPermissions {
			if strings.ToLower(x) == "public" {
				perms = append(perms, types.VisibilitySetting{
					Type:     types.Public,
					Audience: "",
				})
			} else {
				uid, err := resolveUser(x, client)
				if err != nil {
					fmt.Printf("Could not resolve user email or display name: %s\n", x)
					return
				}
				perms = append(perms, types.VisibilitySetting{
					Type:     types.SingleUser,
					Audience: uid,
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
