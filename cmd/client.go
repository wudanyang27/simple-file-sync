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
	Short: "client for simple file sync",
	Long: `A client for simple file sync. For example:
 	simple-file-sync client --local-dir=/Users/wudanyang/self/simple-file-sync --mode=all --remote-dir=/Users/wudanyang/self/testforsimple --server-addr=http://127.0.0.1:8120/receiver --server-token=something`,
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
