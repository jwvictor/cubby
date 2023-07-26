package main

import (
	"fmt"
	"github.com/jwvictor/cubby/pkg/types"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"log"
)

var setCmd = &cobra.Command{
	Use:   "set",
	Short: "Sets the body of a Cubby blob to a new value",
	Long:  `Sets the body of a Cubby blob to a new value by specifying either the ID, the path, or a substring of its title.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			fmt.Printf("Must pass the key, which can be an ID, path, or title substring. New body should be supplied via STDIN.")
			return
		}
		client := getClient()
		err := client.Authenticate()
		if err != nil {
			log.Printf("Error: Authentication - %s\n", err.Error())
			return
		}
		blob, err := client.GetBlobById(args[0])
		if blob == nil {
			log.Printf("No such blob: %s\n", args[0])
			return
		}
		relData := blob.Data
		var encBody *types.BlobBinaryAttachment

		encBody = getBlobEncryptedBody(blob)
		if viper.GetString(CfgEncryptionMode) == CfgEncryptionSymmetric && encBody != nil {
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

		// Overwrite body
		relData = string(slurpStdin())

		newTags := extractTags(relData)
		blob.Tags = append(blob.Tags, newTags...)
		blob.Tags = deduplicateTags(blob.Tags)

		// Now update
		if viper.GetString(CfgEncryptionMode) == CfgEncryptionSymmetric && encBody != nil {
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
