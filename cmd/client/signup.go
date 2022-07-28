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
	//return fmt.Sprintf("%s%d", prefix, rand.Intn(10000)%9999)
	return prefix
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
		dispReq := viper.GetString(CfgUserDisplayName)
		numDisplayNameTries := 5

		var disp string
		if dispReq == "" {
			disp = generateDisplayName(viper.GetString(CfgUserEmail))
			fmt.Printf("No display name provided. Generating random display name from email address %s: %s\n", viper.GetString(CfgUserEmail), disp)
		} else {
			disp = dispReq
		}

		var lastErr error
		for i := 0; i < numDisplayNameTries; i++ {
			if i > 1 && dispReq != "" {
				// We couldn't get a requested display name
				fmt.Printf("We could not sign up that display name with error: %s\n", lastErr.Error())
				os.Exit(1)
			}
			dispTry := disp
			if i > 1 {
				// Add some numbers
				dispTry = fmt.Sprintf("%s%d", disp, rand.Intn(1000))
			}
			err := client.SignUp(dispTry)
			if err != nil {
				if strings.Contains(err.Error(), "Display name already exists") {
					lastErr = err
				} else {
					fmt.Printf("Failed to sign up due to error: %s\n", err.Error())
					os.Exit(1)
				}
			} else {
				fmt.Printf("Signup successful! Your display name is: %s\n", dispTry)
				os.Exit(0)
			}
		}

		// If we got here, we failed
		fmt.Printf("We failed to get a display name similar to your email address. Please pass a display name with the -D flag, e.g.:\n\tcubby signup -D my-display-name\n")
	},
}
