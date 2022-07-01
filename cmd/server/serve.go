package main

import (
	"github.com/jwvictor/cubby/internal/server"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Serve a Cubby API",
	Long:  `Runs a Cubby API service`,
	Run: func(cmd *cobra.Command, args []string) {
		portNum := viper.GetInt("port")
		adminPass := viper.GetString(CfgAdminPass)
		if adminPass == "" {
			panic("Admin pass cannot be empty")
		}
		s := server.NewServer(portNum, adminPass)
		s.Serve()
	},
}
