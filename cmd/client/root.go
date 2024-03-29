package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	CfgPort            = "port"
	CfgHost            = "host"
	CfgUserEmail       = "user.email"
	CfgUserPassword    = "user.password"
	CfgUserDisplayName = "user.display-name"
	CfgConfigFile      = "config"
	CfgViewer          = "options.viewer"
	CfgBodyOnly        = "options.body-only"
	CfgForce           = "options.force"
	CfgSymmetricKey    = "crypto.symmetric-key"
	CfgEncryptionMode  = "crypto.mode"

	CfgViewerStdout = "stdout"
	CfgViewerTui    = "tui"
	CfgViewerEditor = "editor"

	CfgEncryptionNone      = "none"
	CfgEncryptionSymmetric = "symmetric"
)

var (
	cfgFile          string = ""
	host             string = ""
	userEmail        string = ""
	userDisplayName  string = ""
	userPass         string = ""
	userViewer       string = ""
	userSymmetricKey string = ""
	bodyOnly         bool   = false
	useEncryption           = ""
	portNum          int
)

var rootCmd = &cobra.Command{
	Use:   "cubby",
	Short: "Cubby is a personal organizer for the technically inclined",
	Long: `A client-server personal organizer built with love by Jason Victor in Go.
Complete documentation is available at http://www.cubby.com`,
	Run: func(cmd *cobra.Command, args []string) {
		// Do Stuff Here
		fmt.Printf("Use subcommand help for documentation.\n")
	},
}

/**
 libraries for client:
https://github.com/c-bata/go-prompt
https://github.com/rivo/tview
https://github.com/olekukonko/tablewriter
*/

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, CfgConfigFile, "", "config file (default is $HOME/.cobra.yaml)")
	rootCmd.PersistentFlags().IntVarP(&portNum, CfgPort, "p", 80, "example: -p 6969")
	rootCmd.PersistentFlags().StringVarP(&host, CfgHost, "H", "http://public.cubbycli.com", "usually http://public.cubbycli.com")
	rootCmd.PersistentFlags().StringVarP(&userEmail, CfgUserEmail, "e", "me@domain.com", "email associated with user account")
	rootCmd.PersistentFlags().StringVarP(&userPass, CfgUserPassword, "P", "s3cre7", "password associated with user account")
	rootCmd.PersistentFlags().StringVarP(&userDisplayName, CfgUserDisplayName, "D", "Your display name", "Display name to associate with shares, etc.")
	rootCmd.PersistentFlags().StringVarP(&userViewer, CfgViewer, "V", CfgViewerStdout, "one of viewer, stdout, editor")
	rootCmd.PersistentFlags().BoolVarP(&bodyOnly, CfgBodyOnly, "b", false, "show body only when displaying data")
	rootCmd.PersistentFlags().BoolVarP(&force, CfgForce, "f", false, "force a command to skip validation")
	rootCmd.PersistentFlags().StringVarP(&userSymmetricKey, CfgSymmetricKey, "K", "", "a passphrase to use as a symmetric encryption key")
	rootCmd.PersistentFlags().StringVarP(&useEncryption, CfgEncryptionMode, "C", "", "encryption mode")
	viper.BindPFlag(CfgPort, rootCmd.PersistentFlags().Lookup(CfgPort))
	viper.BindPFlag(CfgHost, rootCmd.PersistentFlags().Lookup(CfgHost))
	viper.BindPFlag(CfgUserEmail, rootCmd.PersistentFlags().Lookup(CfgUserEmail))
	viper.BindPFlag(CfgUserPassword, rootCmd.PersistentFlags().Lookup(CfgUserPassword))
	viper.BindPFlag(CfgUserDisplayName, rootCmd.PersistentFlags().Lookup(CfgUserDisplayName))
	viper.BindPFlag(CfgViewer, rootCmd.PersistentFlags().Lookup(CfgViewer))
	viper.BindPFlag(CfgBodyOnly, rootCmd.PersistentFlags().Lookup(CfgBodyOnly))
	viper.BindPFlag(CfgSymmetricKey, rootCmd.PersistentFlags().Lookup(CfgSymmetricKey))
	viper.BindPFlag(CfgEncryptionMode, rootCmd.PersistentFlags().Lookup(CfgEncryptionMode))
	viper.SetDefault(CfgPort, 8888)
	viper.SetDefault(CfgViewer, CfgViewerStdout)
	viper.SetDefault(CfgBodyOnly, false)
	viper.SetDefault(CfgUserDisplayName, "")

	putCmd.Flags().StringVarP(&putCmdTitle, "title", "t", "", "a title for my blob")
	putCmd.Flags().StringVarP(&putCmdType, "type", "T", "markdown", "example: todo")
	putCmd.Flags().StringVarP(&putCmdData, "data", "d", "", "example: *some* markdown")
	putCmd.Flags().StringVarP(&putCmdTags, "tags", "g", "", "example: work urgent")
	putCmd.Flags().StringVarP(&putCmdParentId, "parent", "1", "", "parent UUID")
	putCmd.Flags().StringVarP(&putCmdParentPath, "parentPath", "2", "", "path/to/parent")
	putCmd.Flags().IntVarP(&putCmdImportance, "importance", "I", 0, "example: 0")
	putCmd.Flags().StringArrayVarP(&putCmdAttachFilenames, "attachment", "a", nil, "files to attach")
	putCmd.Flags().DurationVarP(&putCmdTtl, "ttl", "X", time.Duration(0), "optional TTL (0 for no TTL)")

	getPublicationCmd.Flags().StringVarP(&publishOwnerId, "publish.ownerId", "O", "", "Display name of the owner of a shared post (or empty yourself)")
	putPublicationCmd.Flags().StringVarP(&publishPostId, "publish.postId", "Z", "", "ID for the post - can be any alphanumeric string (defaults to title)")
	putPublicationCmd.Flags().StringVarP(&publishPublicationId, "publish.publicationId", "z", "", "ID for the publication this post post will belong to (optional)")
	putPublicationCmd.Flags().StringArrayVarP(&publishPermissions, "publish.permissions", "r", nil, "Permissions for the post - can be `public` or a user ID")

	attachmentsCmd.Flags().StringArrayVarP(&attachmentCmdFiles, "files", "F", nil, "files to download")

	grepCmd.Flags().BoolVarP(&grepCaseInsensitive, "case-insensitive", "i", false, "toggle case insensitivity")

  listCmd.Flags().BoolVarP(&listCmdRenderJson, "json-listing", "J", false, "get listing as JSON, e.g. for neovim plugin")

	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(upgradeCmd)
	rootCmd.AddCommand(getCmd)
	rootCmd.AddCommand(catCmd)
	rootCmd.AddCommand(grepCmd)
	rootCmd.AddCommand(revertCmd)
	rootCmd.AddCommand(attachmentsCmd)
	rootCmd.AddCommand(attachCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(signupCmd)
	rootCmd.AddCommand(profileCmd)
	rootCmd.AddCommand(pushCmd)
	rootCmd.AddCommand(setCmd)
	rootCmd.AddCommand(rmCmd)
	rootCmd.AddCommand(searchCmd)
	// Disabling TUI until we can make it better
	//rootCmd.AddCommand(tuiCmd)
	rootCmd.AddCommand(putCmd)

	rootCmd.AddCommand(publishCmd)
	publishCmd.AddCommand(putPublicationCmd)
	publishCmd.AddCommand(rmPublicationCmd)
	publishCmd.AddCommand(getPublicationCmd)
	publishCmd.AddCommand(listPublicationsCmd)

	rootCmd.AddCommand(adminCmd)
	adminCmd.AddCommand(adminStatsCmd)

	profileCmd.AddCommand(profileSearchCmd)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		cubbyDir := filepath.Join(home, ".cubby")

		// Search config in home directory with name ".cobra" (without extension).
		viper.AddConfigPath(".")
		viper.AddConfigPath(home)
		viper.AddConfigPath(cubbyDir)
		viper.SetConfigType("yaml")
		viper.SetConfigName("cubby-client")
	}

  viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
	viper.SetEnvPrefix("cub") // will be uppercased automatically
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		//fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
