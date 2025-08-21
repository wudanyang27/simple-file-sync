/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"log"
	"os"

	"github.com/spf13/cobra"
	"github.com/wudanyang6/simple-file-sync/config"
)

var (
	cfgFile    string
	globalConfig *config.Config
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "simple-file-sync",
	Short: "simple file sync",
	Long:  `a simple file sync from client to server. it just handle add and modify file event.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		initConfig()
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is simple-file-sync.toml in current directory or $HOME/.simple-file-sync.toml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		cfg, err := config.LoadConfig(cfgFile)
		if err != nil {
			log.Fatalf("Error loading config file: %v", err)
		}
		globalConfig = cfg
	} else {
		// Try to find a config file automatically
		if foundConfig, err := config.FindConfigFile(); err == nil {
			cfg, err := config.LoadConfig(foundConfig)
			if err != nil {
				log.Printf("Warning: Found config file but failed to load: %v", err)
				globalConfig = config.DefaultConfig()
			} else {
				globalConfig = cfg
			}
		} else {
			// No config file found, use defaults
			globalConfig = config.DefaultConfig()
		}
	}
}
