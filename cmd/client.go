/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/spf13/cobra"

	"github.com/wudanyang6/simple-file-sync/client"
)

var (
	ClientUploadMode     string
	ClientLocalDir       string
	ClientRemoteDir      string
	ClientServerAddr     string
	ClientServerToken    string
	ClientIgnorePatterns []string
	ClientPathMappings   []string
	ClientConfigFile     string
)

// ClientConfig 表示客户端配置文件结构
type ClientConfig struct {
	Mode         string   `toml:"mode"`
	LocalDir     string   `toml:"local_dir"`
	RemoteDir    string   `toml:"remote_dir"`
	ServerAddr   string   `toml:"server_addr"`
	ServerToken  string   `toml:"server_token"`
	Ignore       []string `toml:"ignore"`
	PathMappings []string `toml:"path_mappings"`
}

// loadConfig 从TOML文件加载配置
func loadConfig(configFile string) (*ClientConfig, error) {
	// 默认配置文件
	if configFile == "" {
		configFile = "simple-file-sync.toml"
	}

	// 检查文件是否存在
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		// 如果明确指定了配置文件但不存在，返回错误
		if ClientConfigFile != "" {
			return nil, fmt.Errorf("配置文件 %s 不存在", configFile)
		}
		// 使用默认配置
		return &ClientConfig{}, nil
	}

	var config ClientConfig
	if _, err := toml.DecodeFile(configFile, &config); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %v", err)
	}

	return &config, nil
}

// overrideConfigWithFlags 使用命令行参数覆盖配置文件设置
func overrideConfigWithFlags(config *ClientConfig) {
	// 只有当命令行参数有值时才覆盖配置文件中的值
	if ClientUploadMode != "" {
		config.Mode = ClientUploadMode
	}
	if ClientLocalDir != "" {
		config.LocalDir = ClientLocalDir
	}
	if ClientRemoteDir != "" {
		config.RemoteDir = ClientRemoteDir
	}
	if ClientServerAddr != "" {
		config.ServerAddr = ClientServerAddr
	}
	if ClientServerToken != "" {
		config.ServerToken = ClientServerToken
	}
	if len(ClientIgnorePatterns) > 0 {
		config.Ignore = append(config.Ignore, ClientIgnorePatterns...)
	}
	if len(ClientPathMappings) > 0 {
		config.PathMappings = append(config.PathMappings, ClientPathMappings...)
	}
}

// validateConfig 验证配置是否有效
func validateConfig(config *ClientConfig) error {
	if config.Mode == "" {
		config.Mode = "all" // 默认模式
	} else if config.Mode != "all" && config.Mode != "git" {
		return fmt.Errorf("不支持的模式: %s，支持的模式: all, git", config.Mode)
	}

	if config.LocalDir == "" {
		return fmt.Errorf("必须指定本地目录 (local_dir)")
	}
	if config.ServerAddr == "" {
		return fmt.Errorf("必须指定服务器地址 (server_addr)")
	}
	if config.RemoteDir == "" {
		return fmt.Errorf("必须指定远程目录 (remote_dir)")
	}

	return nil
}

// clientCmd represents the client command
var clientCmd = &cobra.Command{
	Use:   "client",
	Short: "client for simple file sync",
	Long: `A client for simple file sync. For example:
 	simple-file-sync client --local-dir=/Users/wudanyang/self/simple-file-sync --mode=all --remote-dir=/Users/wudanyang/self/testforsimple --server-addr=http://127.0.0.1:8120/receiver --server-token=something
 	
 	You can specify files to ignore using regular expressions:
 	--ignore="\.tmp$,node_modules,build"
 	
 	You can specify path mappings using regular expressions:
 	--mapping="/local/path(/.*):/remote/path$1,/local/src(/.+):/remote/build$1"
 	
 	For mapping, the source is a regular expression with capture groups, and target can use $1, $2, etc. to reference captured groups.
 	For example, "/local/path(/.*)" will capture everything after "/local/path" in $1, which can then be referenced in the target path.
 	
 	You can also specify a configuration file in TOML format:
 	-c config.toml or --config=config.toml
 	
 	Example configuration file (simple-file-sync.toml):
 	
 	mode = "all"
 	local_dir = "/Users/wudanyang/self/simple-file-sync"
 	remote_dir = "/Users/wudanyang/self/testforsimple"
 	server_addr = "http://127.0.0.1:8120/receiver"
 	server_token = "something"
 	ignore = ["\\.tmp$", "node_modules", "build"]
 	path_mappings = ["/local/path(/.*):remote/path$1"]
 	`,
	Run: func(cmd *cobra.Command, args []string) {
		// 加载配置文件
		config, err := loadConfig(ClientConfigFile)
		if err != nil {
			log.Fatalf("加载配置失败: %v", err)
		}

		// 使用命令行参数覆盖配置文件
		overrideConfigWithFlags(config)

		// 验证配置
		if err := validateConfig(config); err != nil {
			log.Fatalf("配置验证失败: %v", err)
		}

		// 创建客户端
		client := client.NewClient(
			config.Mode,
			config.LocalDir,
			config.ServerAddr,
			config.RemoteDir,
			config.ServerToken,
		)

		// 添加忽略模式
		for _, pattern := range config.Ignore {
			client.AddIgnorePattern(pattern)
		}

		// 添加路径映射
		for _, mapping := range config.PathMappings {
			parts := strings.Split(mapping, ":")
			if len(parts) == 2 {
				client.AddPathMapping(parts[0], parts[1])
			}
		}

		// 默认添加基础目录映射
		if config.LocalDir != "" && config.RemoteDir != "" {
			// 将基础目录作为正则表达式和替换模式，需要转义特殊字符
			source := "^" + regexp.QuoteMeta(config.LocalDir) + "(/.*)?$"
			target := config.RemoteDir + "$1"
			client.AddPathMapping(source, target)
		}

		client.Start()
	},
}

func init() {
	rootCmd.AddCommand(clientCmd)
	clientCmd.Flags().StringVar(&ClientUploadMode, "mode", "", "upload mode, sync or async")
	clientCmd.Flags().StringVar(&ClientLocalDir, "local-dir", "", "local directory")
	clientCmd.Flags().StringVar(&ClientRemoteDir, "remote-dir", "", "remote directory")
	clientCmd.Flags().StringVar(&ClientServerAddr, "server-addr", "", "server address")
	clientCmd.Flags().StringVar(&ClientServerToken, "server-token", "", "server token")

	// 添加ignore和mapping的flags
	clientCmd.Flags().StringSliceVar(&ClientIgnorePatterns, "ignore", []string{}, "regex patterns to ignore, comma separated, e.g. \\.tmp$,node_modules")
	clientCmd.Flags().StringSliceVar(&ClientPathMappings, "mapping", []string{}, "path mappings using regex, format: source-regex:target-template, comma separated, e.g. /local/path(/.*):remote/path$1")

	// 添加配置文件选项
	clientCmd.Flags().StringVarP(&ClientConfigFile, "config", "c", "", "configuration file path (default \"simple-file-sync.toml\")")
}
