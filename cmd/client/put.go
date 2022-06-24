package main

import (
	"fmt"
	"github.com/jwvictor/cubby/pkg/types"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"io/ioutil"
	"log"
	"path/filepath"
	"strings"
	"time"
)

var (
	putCmdTitle                    = ""
	putCmdType                     = ""
	putCmdData                     = ""
	putCmdParentId                 = ""
	putCmdParentPath               = ""
	putCmdImportance               = 0
	putCmdTags                     = ""
	putCmdTtl                      = time.Duration(0)
	putCmdAttachFilenames []string = nil
)

var putCmd = &cobra.Command{
	Use:   "put",
	Short: "Put a blob in Cubby",
	Long:  `Puts a blob in Cubby based on command line flags set.`,
	Run: func(cmd *cobra.Command, args []string) {
		client := getClient()
		client.CheckVersions()
		err := client.Authenticate()
		if err != nil {
			log.Printf("Error: Authentication - %s\n", err.Error())
		}
		var tags []string = nil
		if putCmdTags != "" {
			tags = strings.Split(putCmdTags, " ")
		}
		var parentBlobId string
		if putCmdParentPath != "" {
			parentBlob, err := client.GetBlobById(putCmdParentPath)
			if err != nil {
				log.Printf("Error: Could not resolve parent `%s`\n", putCmdParentPath)
				return
			} else {
				parentBlobId = parentBlob.Id
				log.Printf("Setting parent ID to %s...\n", parentBlob.Id)
			}
		} else if putCmdParentId != "" {
			parentBlobId = putCmdParentId
		}
		if putCmdData == "" && putCmdTitle == "" {
			// Using args
			if len(args) < 2 {
				putCmdData = strings.Join(args, " ")
				putCmdTitle = putCmdData
				if len(putCmdTitle) > 32 {
					putCmdTitle = putCmdTitle[:32]
				}
			} else {
				putCmdTitle = args[0]
				putCmdData = strings.Join(args[1:], " ")
			}
		} else if putCmdData != "" || putCmdTitle != "" {
			// Not allowed
			fmt.Printf("Please use either command-line flags or arguments, but not both. Examples:\n\tcubby put -t title -d data\n\tcubby put title data\n")
			return
		}
		if putCmdData == "" {
			// Then try using args
			putCmdData = strings.Join(args, " ")
		}
		if putCmdTitle == "" {
			if len(putCmdData) < 32 {
				putCmdTitle = putCmdData
			} else {
				putCmdTitle = putCmdData[:32]
			}
		}
		if putCmdTitle == "" && putCmdData == "" {
			fmt.Errorf("Cannot put an empty blob.\n")
			return
		}
		blob := &types.Blob{
			Title:      putCmdTitle,
			ParentId:   parentBlobId,
			Type:       putCmdType,
			Data:       putCmdData,
			Importance: putCmdImportance,
			Tags:       tags,
			Deleted:    false,
		}

		newTags := extractTags(putCmdData)
		blob.Tags = append(blob.Tags, newTags...)
		blob.Tags = deduplicateTags(blob.Tags)

		if viper.GetString(CfgEncryptionMode) == CfgEncryptionSymmetric {
			key, err := types.DeriveSymmetricKey(viper.GetString(CfgSymmetricKey))
			if err != nil {
				log.Printf("Error: %s\n", err.Error())
				return
			}
			cipherTxt, err := types.EncryptSymmetric([]byte(blob.Data), key)
			if err != nil {
				log.Printf("Error: %s\n", err.Error())
				return
			}
			blob.RawData = append(blob.RawData, types.BlobBinaryAttachment{
				Description: "",
				Data:        cipherTxt,
				Type:        types.EncryptedBody,
			})
			blob.Data = ""
		}

		if putCmdTtl != time.Duration(0) {
			expireTime := time.Now().Add(putCmdTtl)
			blob.ExpireTime = &expireTime
		}

		if len(putCmdAttachFilenames) > 0 {
			for _, fn := range putCmdAttachFilenames {
				err := putOrUpdateAttachment(blob, fn)
				if err != nil {
					fmt.Printf("Error adding attachment: %s\n", err.Error())
					return
				}
			}
		}

		id, err := client.PutBlob(blob)
		if err != nil {
			log.Printf("Error: %s\n", err.Error())
		} else {
			log.Printf("Done (%s).\n", id)
		}
	},
}

func putOrUpdateAttachment(blob *types.Blob, filename string) error {
	fileBase := filepath.Base(filename)
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	useEncryption := false
	var symmKey, cipherTxt []byte
	if viper.GetString(CfgEncryptionMode) == CfgEncryptionSymmetric {
		useEncryption = true
		key, err := types.DeriveSymmetricKey(viper.GetString(CfgSymmetricKey))
		if err != nil {
			return err
		} else {
			symmKey = key
		}
		cipherTxt, err = types.EncryptSymmetric([]byte(data), symmKey)
		if err != nil {
			return err
		}
	}

	for _, item := range blob.RawData {
		if !useEncryption {
			if item.Description == fileBase && item.Type == types.Attachment {
				item.Data = data
				return nil
			}
		} else {
			if item.Description == fileBase && item.Type == types.EncryptedAttachment {
				item.Data = cipherTxt
				return nil
			}
		}
	}

	// If we got here, it's an insert
	if useEncryption {
		blob.RawData = append(blob.RawData, types.BlobBinaryAttachment{
			Description: fileBase,
			Data:        cipherTxt,
			Type:        types.EncryptedAttachment,
		})
	} else {
		blob.RawData = append(blob.RawData, types.BlobBinaryAttachment{
			Description: fileBase,
			Data:        data,
			Type:        types.Attachment,
		})
	}

	return nil
}
