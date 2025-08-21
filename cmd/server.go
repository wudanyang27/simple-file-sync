/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"

	"github.com/wudanyang6/simple-file-sync/server"
)

var (
	ServerPort     int
	ServerToken    string
	ServerLimitDir string
)

// serverCmd represents the server command
var serverCmd = &cobra.Command{
	Use:   "server --port",
	Short: "server for simple file sync",
	Long:  `A server for simple file sync. For example: simple-file-sync server `,
	Run: func(cmd *cobra.Command, args []string) {
		server.NewServer(ServerPort, ServerToken, ServerLimitDir).Start()
	},
}

func init() {
	rootCmd.AddCommand(serverCmd)

	serverCmd.Flags().IntVar(&ServerPort, "port", 8120, "port to listen on")
	serverCmd.Flags().StringVar(&ServerToken, "token", defaultServerToken, "token for authentication")

	homeDir, err := homedir.Dir()
	if err != nil {
		// 未找到家目录，使用/tmp
		homeDir = "/tmp"
	}
	serverCmd.Flags().StringVar(&ServerLimitDir, "limit-dir", "/your/homeDir", "You can’t upload to anything other than the limit-dir folder, default is your home directory")
	ServerLimitDir = homeDir
}
