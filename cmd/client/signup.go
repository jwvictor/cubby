package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"math/rand"
	"os"
	"strings"
	"time"
)

func generateDisplayName(email string) string {
	lidx := strings.Index(email, "@")
	if lidx < 0 {
		lidx = len(email)
	}
	prefix := email[:lidx]
	return fmt.Sprintf("%s%d", prefix, rand.Intn(10000)%9999)
}

var profileCmd = &cobra.Command{
	Use:   "profile",
	Short: "Show user profile",
	Long:  `Show user profile, such as registered email and display name.`,
	Run: func(cmd *cobra.Command, args []string) {
		client := getClient()
		client.CheckVersions()
		rand.Seed(time.Now().Unix())
		err := client.Authenticate()
		if err != nil {
			fmt.Printf("Unable to get authenticate with error: %s\n\nAre you sure you've signed up?\n", err.Error())
			return
		}
		user, err := client.UserProfile()
		if err != nil {
			fmt.Printf("Unable to get user profile with error: %s\n\nAre you sure you've signed up?\n", err.Error())
			return
		}
		fmt.Printf("Email:            %s\n", user.Email)
		fmt.Printf("Id:               %s\n", user.Id)
		fmt.Printf("Display name:     %s\n", user.DisplayName)
	},
}

var profileSearchCmd = &cobra.Command{
	Use:   "search",
	Short: "Search user profile",
	Long:  `Search user profile by email or display name.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := getClient()
		client.CheckVersions()
		rand.Seed(time.Now().Unix())
		err := client.Authenticate()
		if err != nil {
			fmt.Printf("Unable to get authenticate with error: %s\n\nAre you sure you've signed up?\n", err.Error())
			return
		}
		fmt.Printf("Searching for user with email or display name: %s\n", args[0])
		user, err := client.SearchUser(args[0])
		if err != nil {
			fmt.Printf("Unable to get user profile with error: %s\n", err.Error())
			return
		}
		fmt.Printf("Email:            %s\n", user.Email)
		fmt.Printf("Id:               %s\n", user.Id)
		fmt.Printf("Display name:     %s\n", user.DisplayName)
	},
}

var signupCmd = &cobra.Command{
	Use:   "signup",
	Short: "Sign up a new user account with Cubby",
	Long:  `Signs up a new email address for a new user account.`,
	Run: func(cmd *cobra.Command, args []string) {
		client := getClient()
		client.CheckVersions()
		rand.Seed(time.Now().Unix())
		disp := viper.GetString(CfgUserDisplayName)
		if disp == "" {
			disp = generateDisplayName(viper.GetString(CfgUserEmail))
			fmt.Printf("No display name provided. Generating random display name: %s\n", disp)
		}
		err := client.SignUp(disp)
		if err != nil {
			fmt.Printf("Failed to sign up with error: %s\n", err.Error())
			os.Exit(1)
		} else {
			fmt.Printf("Signup successful!\n")
		}
	},
}
