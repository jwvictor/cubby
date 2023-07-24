package main

import (
	"fmt"
	"github.com/jwvictor/cubby/pkg/types"
	"github.com/spf13/cobra"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of Cubby",
	Long:  `Just a version number, like all version numbers.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Cubby client v%s \n", types.ClientVersion)
		client := getClient()
		client.CheckVersions()
	},
}

var upgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "Upgrade Cubby binary",
	Long:  `Upgrades Cubby binary to latest available version`,
	Run: func(cmd *cobra.Command, args []string) {
		client := getClient()
		_, upgradeUrl, err := client.CheckVersions()
		fmt.Printf("Got upgrade URL %s...\n", upgradeUrl)
		if err != nil {
			fmt.Printf("Error: could not get version information: %s\n", err.Error())
			return
		} else {
			scriptBytes, err := client.FetchInstallScript(upgradeUrl)
			if err != nil {
				fmt.Printf("Error: could not get install script: %s\n", err.Error())
				return
			}
			fn := filepath.Join("/tmp", fmt.Sprintf("cubbyinstall_%d.sh", time.Now().Unix()))
			fmt.Printf("Writing install script to %s...\n", fn)
			err = ioutil.WriteFile(fn, scriptBytes, 0644)
			if err != nil {
				fmt.Printf("Error: could not write install script to `tmp`: %s\nPlease upgrade by following the installation instructions on https://github.com/jwvictor/cubby instead.\n", err.Error())
				return
			}

      // Swallow the error in case the file isn't there or similar
      os.Rename(filepath.Join(os.Getenv("HOME"), ".cubby/bin/cubby"), filepath.Join(os.Getenv("HOME"), ".cubby/bin/cubby.old"))

			cmd := exec.Command("bash", fn)
			cmd.Stdin = os.Stdin
			cmd.Stdout = os.Stdout
			err = cmd.Run()
			if err != nil {
				fmt.Printf("Error running install script: %s\n", err.Error())
				return
			}

			err = os.Remove(fn)
			if err != nil {
				fmt.Printf("Error: could not delete install script at: %s\n", fn)
				return
			}
		}
	},
}
