package main

import (
	"fmt"
	"github.com/jwvictor/cubby/pkg/types"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"io/ioutil"
	"log"
	"strings"
)

var (
	attachmentCmdFiles []string = nil
	force              bool     = false
)

var attachCmd = &cobra.Command{
	Use:   "attach",
	Short: "Attach a file to a blob",
	Long:  `Adds a provided binary file to a blob using.`,
	Args:  cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		blobId := args[0]
		files := args[1:]

		client := getClient()
		client.CheckVersions()
		err := client.Authenticate()
		if err != nil {
			log.Printf("Error: Authentication - %s\n", err.Error())
			return
		}

		blob, err := client.GetBlobById(blobId)
		if err != nil {
			log.Printf("Error: %s\n", err.Error())
			return
		}

		// Check if the encryption modes don't match
		isBlobEnc := getBlobEncryptedBody(blob) != nil
		areWeEnc := viper.GetString(CfgEncryptionMode) != CfgEncryptionNone

		if (isBlobEnc != areWeEnc) && (!force) {
			fmt.Printf("This blob has a different encryption mode than you have configured right now. We recommend having files and blobs use uniform encryption modes. If you still want to do this, pass the -f flag to force.\n")
			return
		}

		for _, fn := range files {
			err := putOrUpdateAttachment(blob, fn)
			if err != nil {
				fmt.Printf("Failed to attach file `%s`: %s\n", fn, err.Error())
				return
			}
		}

		_, err = client.PutBlob(blob)
		if err != nil {
			log.Printf("Error updating blob: %s\n", err.Error())
		}

	},
}

var attachmentsCmd = &cobra.Command{
	Use:   "attachments",
	Short: "Manage attachments to a blob",
	Long:  `List, download, and decrypt file attachments to Cubby blobs.`,
	Args:  cobra.MinimumNArgs(1),
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
		}

		var filesToDl []types.BlobBinaryAttachment
		for _, desc := range attachmentCmdFiles {
			file := getAttachmentByDesc(blob, desc)
			if file == nil {
				fmt.Printf("Error: no such file `%s` in this blob. Available files are: %s\n", desc, strings.Join(getAttachmentDescs(blob), ", "))
				return
			}
			filesToDl = append(filesToDl, *file)
		}

		if len(attachmentCmdFiles) == 0 {
			// Just print the files that exist
			names := getAttachmentDescs(blob)
			for i, n := range names {
				fmt.Printf(" %d. %s\n", i+1, n)
			}
			if len(names) == 0 {
				fmt.Printf("No attachments to this blob.\n")
			}
			return
		}

		for _, file := range filesToDl {
			if file.Type == types.EncryptedAttachment {
				plainBody, err := decryptData(file.Data)
				if err != nil {
					fmt.Printf("Error decrypting blob: %s\n", err.Error())
					return
				}
				err = outputFile(plainBody, file.Description)
				if err != nil {
					fmt.Printf("Error writing file `%s`: %s\n", file.Description, err.Error())
					return
				}
			} else {
				err := outputFile(file.Data, file.Description)
				if err != nil {
					fmt.Printf("Error writing file `%s`: %s\n", file.Description, err.Error())
					return
				} else {
					fmt.Printf("Successfully wrote file: %s\n", file.Description)
				}
			}
		}
	},
}

func outputFile(data []byte, filename string) error {
	return ioutil.WriteFile(filename, data, 0644)
}
