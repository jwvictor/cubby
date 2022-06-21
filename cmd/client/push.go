package main

import (
	"fmt"
	"github.com/jwvictor/cubby/pkg/types"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"log"
	"strings"
)

var pushCmd = &cobra.Command{
	Use:   "push",
	Short: "Add a line of text to a blob from Cubby",
	Long:  `Adds a new line of text to a blob from Cubby by specifying either its ID, path, or a substring of its title.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 2 {
			fmt.Printf("Must pass two arguments: (i.) the key, which can be an ID, path, or title substring, and (ii.) the line of text to add.")
			return
		}
		client := getClient()
		err := client.Authenticate()
		if err != nil {
			log.Printf("Error: Authentication - %s\n", err.Error())
			return
		}
		blob, err := client.GetBlobById(args[0])

		relData := blob.Data
		var encBody *types.BlobBinaryAttachment

		if viper.GetString(CfgEncryptionMode) == CfgEncryptionSymmetric {
			encBody = getBlobEncryptedBody(blob)
			if encBody == nil {
				log.Printf("Error: cannot find encrypted body\n")
				return
			}
			relDataBs, err := decryptData(encBody.Data)
			if err != nil {
				log.Printf("Error: %s\n", err.Error())
				return
			}
			relData = string(relDataBs)
		}

		if !strings.HasSuffix(relData, "\n") {
			// Add a newline
			relData += "\n"
		}
		for i := 1; i < len(args); i++ {
			relData += args[1]
		}

		newTags := extractTags(relData)
		blob.Tags = append(blob.Tags, newTags...)
		blob.Tags = deduplicateTags(blob.Tags)

		// Now update
		if viper.GetString(CfgEncryptionMode) == CfgEncryptionSymmetric {
			key, err := types.DeriveSymmetricKey(viper.GetString(CfgSymmetricKey))
			if err != nil {
				fmt.Errorf("Error deriving key: %s\n", err.Error())
				return
			}
			newEncData, err := types.EncryptSymmetric([]byte(relData), key)
			if err != nil {
				fmt.Errorf("Error encrypting data: %s\n", err.Error())
				return
			}
			for i, x := range blob.RawData {
				if x.Type == types.EncryptedBody {
					blob.RawData[i].Data = newEncData
				}
			}
		} else {
			blob.Data = relData
		}

		_, err = client.PutBlob(blob)
		if err != nil {
			fmt.Errorf("Error updating blob: %s\n", err.Error())
		}
	},
}
