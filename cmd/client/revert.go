package main

import (
	"fmt"
	"github.com/jwvictor/cubby/pkg/types"
	"github.com/spf13/cobra"
	"log"
	"strconv"
	"strings"
)

var revertCmd = &cobra.Command{
	Use:   "revert blob-path",
	Short: "Revert a blob",
	Long:  `Reverts a blob to a prior version by specifying either its ID, path, or a substring of its title.`,
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
			log.Printf("Error getting blob: %s\n", err.Error())
			return
		} else {
			histItem := displayVersionsAndAsk(blob)

			if histItem != nil {
				if encBody := getBlobEncryptedBody(histItem.BlobValue); encBody != nil {
					bs, err := decryptData(encBody.Data)
					if err != nil {
						fmt.Errorf("Error decrypting: %s\n", err.Error())
						return
					} else {
						fmt.Printf("Will replace contents with: %s\n\n\n", string(bs))
					}
				} else {
					fmt.Printf("Will replace contents with: %s\n\n\n", histItem.BlobValue.Data)
				}
				stdin, err := readFromStdin("Proceed with revert? (y/N)")
				if err != nil {
					fmt.Printf("Invalid input. Please try again.\n")
					return
				}
				stdin = strings.ToLower(stdin)
				if strings.HasPrefix(stdin, "y") {
					// Do revert
					if encBody := getBlobEncryptedBody(histItem.BlobValue); encBody != nil {
						for i, x := range blob.RawData {
							if x.Type == types.EncryptedBody {
								blob.RawData[i].Data = encBody.Data
							}
						}
					} else {
						blob.Data = histItem.BlobValue.Data
					}
					_, err := client.PutBlob(blob)
					if err != nil {
						fmt.Printf("Failed to update blob: %s\n", err.Error())
					}
				}
			}
		}
	},
}

func displayVersionsAndAsk(blob *types.Blob) *types.BlobHistoryItem {
	for i := 1; i <= 30; i++ {
		desc := blob.Data
		if i >= len(blob.VersionHistory) {
			break
		}
		rig := blob.VersionHistory[len(blob.VersionHistory)-i]
		var err error
		var bs []byte
		encBody := getBlobEncryptedBody(rig.BlobValue)
		if encBody != nil {
			bs, err = decryptData(encBody.Data)
			if err != nil {
				fmt.Errorf("Error decrypting: %s\n", err.Error())
				return nil
			}
			desc = string(bs)
		}
		desc = strings.ReplaceAll(desc, "\n", " ")
		maxChars := 128
		if len(desc) > maxChars {
			desc = desc[:maxChars]
		}
		fmt.Printf("[%d]\t%s\t%s\n", i, rig.AsOf.String(), desc)
	}

	for {
		stdin, err := readFromStdin("Which version do you want to revert to?")
		if err != nil {
			fmt.Printf("Invalid input. Please try again.\n")
			continue
		}
		iv, err := strconv.ParseInt(stdin, 10, 64)
		if err != nil {
			fmt.Printf("Invalid input. Please try again.\n")
			continue
		}
		if int(iv) >= len(blob.VersionHistory) {
			fmt.Printf("Invalid input. Please try again.\n")
			continue
		}
		histItem := blob.VersionHistory[len(blob.VersionHistory)-int(iv)]
		return histItem
	}
}
