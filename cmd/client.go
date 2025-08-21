/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"log"

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
		// Use config values as defaults, but allow CLI flags to override
		mode := ClientUploadMode
		localDir := ClientLocalDir
		remoteDir := ClientRemoteDir
		serverAddr := ClientServerAddr
		serverToken := ClientServerToken

		// Apply config values if not explicitly set via flags
		if globalConfig != nil {
			if mode == "all" && globalConfig.Mode != "" {
				mode = globalConfig.Mode
			}
			if localDir == "" && globalConfig.LocalDir != "" {
				localDir = globalConfig.LocalDir
			}
			if remoteDir == "" && globalConfig.RemoteDir != "" {
				remoteDir = globalConfig.RemoteDir
			}
			if serverAddr == "" && globalConfig.ServerAddr != "" {
				serverAddr = globalConfig.ServerAddr
			}
			if serverToken == defaultServerToken && globalConfig.ServerToken != "" {
				serverToken = globalConfig.ServerToken
			}
		}

		// Validate required parameters
		if localDir == "" {
			log.Fatal("local-dir is required (either via --local-dir flag or config file)")
		}
		if serverAddr == "" {
			log.Fatal("server-addr is required (either via --server-addr flag or config file)")
		}
		if remoteDir == "" {
			log.Fatal("remote-dir is required (either via --remote-dir flag or config file)")
		}

		client.NewClient(
			mode,
			localDir,
			serverAddr,
			remoteDir,
			serverToken,
		).Start()
	},
}

func init() {
	rootCmd.AddCommand(clientCmd)
	clientCmd.Flags().StringVar(&ClientUploadMode, "mode", "all", "upload mode, sync or async")
	clientCmd.Flags().StringVar(&ClientLocalDir, "local-dir", "", "local directory")
	clientCmd.Flags().StringVar(&ClientRemoteDir, "remote-dir", "", "remote directory")
	clientCmd.Flags().StringVar(&ClientServerAddr, "server-addr", "", "server address")
	clientCmd.Flags().StringVar(&ClientServerToken, "server-token", defaultServerToken, "server token")
}
