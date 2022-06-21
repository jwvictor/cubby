package main

import (
	"fmt"
	"github.com/spf13/cobra"
)

const (
	Version = "0.0.1"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of Cubby",
	Long:  `Just a version number, like all version numbers.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Cubby server v" + Version)
	},
}

