/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/spf13/cobra"

	"github.com/wudanyang6/simple-file-sync/client"
)

var (
	ClientUploadMode  string
	ClientLocalDir    string
	ClientRemoteDir   string
	ClientServerAddr  string
	ClientServerToken string
)

// clientCmd represents the client command
var clientCmd = &cobra.Command{
	Use:   "client",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		client.NewClient(
			ClientUploadMode,
			ClientLocalDir,
			ClientServerAddr,
			ClientRemoteDir,
			ClientServerToken,
		).Start()
	},
}

func init() {
	rootCmd.AddCommand(clientCmd)
	clientCmd.Flags().StringVar(&ClientUploadMode, "mode", "all", "upload mode, sync or async")
	clientCmd.Flags().StringVar(&ClientLocalDir, "local-dir", "", "local directory")
	clientCmd.Flags().StringVar(&ClientRemoteDir, "remote-dir", "", "remote directory")
	clientCmd.Flags().StringVar(&ClientServerAddr, "server-addr", "", "server address")
	clientCmd.Flags().StringVar(&ClientServerToken, "server-token", serverToken, "server token")
}
