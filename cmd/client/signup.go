package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"math/rand"
	"strings"
	"time"
)

func generateDisplayName(email string) string {
	rand.Seed(time.Now().Unix())
	lidx := strings.Index(email, "@")
	if lidx < 0 {
		lidx = len(email)
	}
	prefix := email[:lidx]
	return fmt.Sprintf("%s%d", prefix, rand.Intn(10000)%9999)
}

var signupCmd = &cobra.Command{
	Use:   "signup",
	Short: "Sign up a new user account with Cubby",
	Long:  `Signs up a new email address for a new user account.`,
	Run: func(cmd *cobra.Command, args []string) {
		client := getClient()
		client.CheckVersions()
		disp := viper.GetString(CfgUserDisplayName)
		if disp == "" {
			disp = generateDisplayName(viper.GetString(CfgUserEmail))
			fmt.Printf("No display name provided. Generating random display name: %s\n", disp)
		}
		err := client.SignUp(disp)
		if err != nil {
			fmt.Printf("Failed to sign up with error: %s\n", err.Error())
		}
	},
}
