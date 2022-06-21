package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
)

const (
	CfgPort       = "port"
	CfgConfigFile = "config"
)

var (
	cfgFile string = ""
	portNum int
)

var rootCmd = &cobra.Command{
	Use:   "cubby",
	Short: "Cubby is a personal organizer for the technically inclined",
	Long: `A client-server personal organizer built with
                love by Jason Victor in Go.
                Complete documentation is available at http://www.cubby.com`,
	Run: func(cmd *cobra.Command, args []string) {
		// Do Stuff Here
		fmt.Printf("Hello\n")
	},
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, CfgConfigFile, "", "config file (default is $HOME/.cobra.yaml)")
	rootCmd.PersistentFlags().IntVarP(&portNum, CfgPort, "p", 8080, "example: -p 6969")
	viper.BindPFlag(CfgPort, rootCmd.PersistentFlags().Lookup(CfgPort))
	viper.SetDefault(CfgPort, 8888)
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(serveCmd)
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

		// Search config in home directory with name ".cobra" (without extension).
		viper.AddConfigPath(home)
		viper.AddConfigPath(".")
		viper.SetConfigType("yaml")
		viper.SetConfigName("cubby-server")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
