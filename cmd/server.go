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

const serverToken = "kfcvme50"

// serverCmd represents the server command
var serverCmd = &cobra.Command{
	Use:   "server --port [args]",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		server.NewServer(ServerPort, ServerToken, ServerLimitDir).Start()
	},
}

func init() {
	rootCmd.AddCommand(serverCmd)

	serverCmd.Flags().IntVar(&ServerPort, "port", 8120, "port to listen on")
	serverCmd.Flags().StringVar(&ServerToken, "token", serverToken, "token for authentication")

	homeDir, err := homedir.Dir()
	if err != nil {
		// 未找到家目录，使用/tmp
		homeDir = "/tmp"
	}
	serverCmd.Flags().StringVar(&ServerLimitDir, "limit-dir", homeDir, "limit directory, start with home dir")
}
