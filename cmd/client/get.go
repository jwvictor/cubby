package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jwvictor/cubby/cmd/client/tuiviewer"
	"github.com/jwvictor/cubby/pkg/client"
	"github.com/jwvictor/cubby/pkg/types"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"time"
	"unicode"
)

const (
	ViewerLoopHelpText = "\nEnter either 1.) a number to view a result, 2.) 'x' followed by a number to delete a result, or 3.) 'q' to quit.\nExamples:\n\tView item #2:\t2\n\tDelete item #2:\tx2\n\n"
)

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Get a blob from Cubby",
	Long:  `Gets a blob from Cubby by specifying either its ID, path, or a substring of its title.`,
	Run: func(cmd *cobra.Command, args []string) {
		client := getClient()
		client.CheckVersions()
		rand.Seed(time.Now().Unix())
		err := client.Authenticate()
		if err != nil {
			log.Printf("Error: Authentication - %s\n", err.Error())
			return
		}
		blob, err := client.GetBlobById(args[0])
		if err != nil {
			log.Printf("Error: %s\n", err.Error())
			return
		} else {
			displayBlob(blob, client, nil, nil)
		}
	},
}

func getBlobEncryptedBody(blob *types.Blob) *types.BlobBinaryAttachment {
	for _, x := range blob.RawData {
		if x.Type == types.EncryptedBody {
			return &x
		}
	}
	return nil
}

func decryptData(data []byte) ([]byte, error) {
	symmKey := viper.GetString(CfgSymmetricKey)
	if symmKey == "" {
		return nil, fmt.Errorf("A symmetric key is required to decrypt this blob.\n")
	}
	key, err := types.DeriveSymmetricKey(symmKey)
	if err != nil {
		return nil, fmt.Errorf("Error generating key: %s\n", err.Error())
	}
	plainBody, err := types.DecryptSymmetric(data, key)
	if err != nil {
		return nil, fmt.Errorf("Error decrypting blob: %s\n", err.Error())
	}
	return plainBody, nil
}

func hasChanged(blob *types.Blob, relData string, client *client.CubbyClient) (bool, error) {
	updatedBlob, err := client.GetBlobById(blob.Id)
	if err != nil {
		return false, err
	}

	updatedEncBody := getBlobEncryptedBody(updatedBlob)
	updatedRelData, done := extractRelData(updatedBlob, updatedEncBody)
	if done {
		return false, errors.New("DecryptionFailed")
	}

	if updatedRelData != relData {
		return true, nil
	} else {
		return false, nil
	}
}

func displayBlob(blob *types.Blob, client *client.CubbyClient, viewerOverride *string, bodyOnlyOverride *bool) {
	bs, err := json.MarshalIndent(blob, "", "    ")
	if err != nil {
		log.Printf("Error: %s\n", err.Error())
	}
	bodyOnly := viper.GetBool(CfgBodyOnly)
	userViewer := viper.GetString(CfgViewer)
	encBody := getBlobEncryptedBody(blob)

	relData, done := extractRelData(blob, encBody)
	if done {
		return
	}

	if viewerOverride != nil {
		userViewer = *viewerOverride
	}

	if bodyOnlyOverride != nil {
		bodyOnly = *bodyOnlyOverride
	}

	switch userViewer {
	case CfgViewerTui:
		tuiviewer.RunViewer(relData, blob.Title)
	case CfgViewerEditor:
		fileext := ""
		contentType := types.ResolveContentType(blob.Type)
		if contentType != nil {
			fileext = contentType.FileExtension
		}
		newData, err := openInEditor(relData, blob.Title, fileext)
		if err != nil {
			fmt.Errorf("Failed to open editor: %s\n", err.Error())
		}

		if newData != relData {
			if changed, err := hasChanged(blob, relData, client); err == nil {
				if changed {
					// Ask if user really wants to save it
					v, err := readFromStdin("This blob has changed since you started editing it. Are you sure you want to overwrite it? (y/N) ")
					if err != nil {
						fmt.Printf("Failed to read from stdin, quitting...\n")
						return
					} else {
						vl := strings.ToLower(v)
						if strings.HasPrefix(vl, "y") {
							// yes
						} else {
							// no
							fmt.Printf("Aborting save. Printing data to STDOUT in case you want to save it some other way:\n\n%s\n", newData)
							return
						}
					}
				}
			} else {
				// failed to detect changes
				fmt.Printf("Failed to fetch old version in order to run overwrite protection checks: aborting save. Printing data to STDOUT in case you want to save it some other way:\n\n%s\n", newData)
			}
			newTags := extractTags(newData)
			blob.Tags = append(blob.Tags, newTags...)
			blob.Tags = deduplicateTags(blob.Tags)

			// User updated blob
			if getBlobEncryptedBody(blob) != nil {
				// Encrypted case
				key, err := types.DeriveSymmetricKey(viper.GetString(CfgSymmetricKey))
				if err != nil {
					fmt.Errorf("Error deriving key: %s\n", err.Error())
					return
				}
				newEncData, err := types.EncryptSymmetric([]byte(newData), key)
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
				// Not encrypted case
				blob.Data = newData
			}

			blob.VersionHistory = nil

			//fmt.Printf("PUTTING THIS : \n%+v\n\n", blob)
			_, err := client.PutBlob(blob)
			if err != nil {
				fmt.Printf("Failed to update blob: %s\n", err.Error())
			}
		}
	case CfgViewerStdout:
		if bodyOnly {
			fmt.Printf("%s\n", relData)
		} else {
			fmt.Printf("%s\n", string(bs))
		}
	}
}

func extractRelData(blob *types.Blob, encBody *types.BlobBinaryAttachment) (string, bool) {
	relData := blob.Data

	if encBody != nil {
		plainBody, err := decryptData(encBody.Data)
		if err != nil {
			fmt.Errorf("Error: %s\n", err.Error())
			return "", true
		}
		relData += string(plainBody)
	}
	return relData, false
}

// allow dashes + underscores in post names

func openInEditor(data, title, fileext string) (string, error) {
	envEditor := os.Getenv("EDITOR")
	var fn string
	for _, x := range title {
		if len(fn) >= 32 {
			break
		}
		if unicode.IsLetter(x) || unicode.IsNumber(x) {
			fn += string(x)
		}
	}
	if len(fn) < 32 {
		fn += fmt.Sprintf("%d", rand.Int()%999999999)
	}
	if fileext != "" {
		fn += "." + fileext
	}
	if envEditor == "" {
		envEditor = "vim"
	}
	fullPath := path.Join("/tmp", fn)
	err := ioutil.WriteFile(fullPath, []byte(data), 0644)
	if err != nil {
		return "", err
	}
	cmd := exec.Command(envEditor, fullPath)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	err = cmd.Run()
	if err != nil {
		return "", err
	}
	newData, err := ioutil.ReadFile(fullPath)
	if err != nil {
		return "", err
	}
	err = os.Remove(fullPath)
	if err != nil {
		return "", err
	}
	return string(newData), nil
}

func readFromStdin(prompt string) (string, error) {
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Printf("%s ", prompt)
	scanner.Scan()

	if err := scanner.Err(); err != nil {
		return "", err
	} else {
		return scanner.Text(), nil
	}
}

func resultsViewer(results []*types.Blob, client *client.CubbyClient) bool {
	ctr := 0
	for {
		ctr += 1
		for i, x := range results {
			fmt.Printf("[%d]\t%s\n", i, x.Title)
		}
		if ctr < 2 {
			fmt.Printf(ViewerLoopHelpText)
		}
		stdin, err := readFromStdin(">")
		if err != nil {
			fmt.Printf("Invalid input. Please try again.\n")
			continue
		}
		stdin = strings.TrimRight(stdin, " \n")
		if stdin == "q" {
			return false
		} else if strings.HasPrefix(stdin, "x") {
			idxStr := stdin[1:]
			idx, err := strconv.ParseInt(idxStr, 10, 64)
			if err != nil {
				fmt.Printf("Invalid number value. Please try again.\n")
				continue
			}
			var blob *types.Blob
			if int(idx) >= len(results) {
				fmt.Printf("Number value out of range. Please try again.\n")
				continue
			} else {
				blob = results[idx]
			}
			res, err := client.DeleteBlob(blob.Id)
			if err != nil {
				fmt.Printf("Error deleting: %s\n", err.Error())
			} else {
				fmt.Printf("Deleted %s.\n", res.Id)
				results = append(results[:idx], results[idx+1:]...)
			}
		} else {
			idx, err := strconv.ParseInt(stdin, 10, 64)
			if err != nil {
				fmt.Printf("Invalid number value. Please try again.\n")
				continue
			}
			var blob *types.Blob
			if int(idx) >= len(results) {
				fmt.Printf("Number value out of range. Please try again.\n")
				continue
			} else {
				blob = results[idx]
			}

			displayBlob(blob, client, nil, nil)
		}
	}
}

var searchCmd = &cobra.Command{
	Use:   "search",
	Short: "Search a blob from Cubby",
	Long:  `Searches for a blob from Cubby by specifying a query string.`,
	Run: func(cmd *cobra.Command, args []string) {
		client := getClient()
		client.CheckVersions()
		rand.Seed(time.Now().Unix())
		err := client.Authenticate()
		if err != nil {
			log.Printf("Error: Authentication - %s\n", err.Error())
		}
		blob, err := client.SearchBlob(strings.Join(args, " "))
		if err != nil {
			log.Printf("Error: %s\n", err.Error())
		} else {
			//bs, err := json.MarshalIndent(blob, "", "    ")
			//if err != nil {
			//	log.Printf("Error: %s\n", err.Error())
			//}
			//fmt.Printf("\n%s\n", string(bs))
			resultsViewer(blob, client)
		}
	},
}

func getAttachmentByDesc(blob *types.Blob, desc string) *types.BlobBinaryAttachment {
	for _, att := range blob.RawData {
		if att.Description == desc {
			return &att
		}
	}
	return nil
}

func getAttachmentDescs(blob *types.Blob) []string {
	var ss []string
	for _, att := range blob.RawData {
		if att.Type == types.EncryptedAttachment || att.Type == types.Attachment {
			ss = append(ss, att.Description)
		}
	}
	return ss
}
