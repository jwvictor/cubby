package main

import (
	"fmt"
	"github.com/jwvictor/cubby/pkg/types"
	"github.com/spf13/cobra"
	"log"
)

var adminCmd = &cobra.Command{
	Use:   "admin",
	Short: "Manage private Cubby servers",
	Long:  `Administrative features for private Cubby servers.`,
}

var adminStatsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Gets server stats",
	Long:  `Gets server stats for a private Cubby server.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := getClient()
		err := client.Authenticate()
		if err != nil {
			log.Printf("Error: Authentication - %s\n", err.Error())
			return
		}

		pass := args[0]
		stats, err := client.AdminStats(pass)
		if err != nil {
			fmt.Printf("Failed to get stats: %s\n", err.Error())
			return
		} else {
			displayStats(stats)
		}
	},
}

func displayStats(res *types.AdminResponse) {
	fmt.Printf("TOTAL USERS:   %d\n", res.NumUsers)
	fmt.Printf("SOME USERS:   \n")
	for _, usr := range res.SomeUsers {
		fmt.Printf("    %s (%s)\n", usr.DisplayName, usr.Email)
	}
}
