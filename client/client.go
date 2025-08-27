package client

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

const (
	NumWorkers = 30
)

type PathMapping struct {
	SourcePattern *regexp.Regexp
	TargetPath    string
}

// RemoteTarget 表示一个远程目标配置
type RemoteTarget struct {
	Name      string
	URL       string
	TargetDir string
	Token     string
}

type Client struct {
	Mode           string
	LocalDir       string
	RemoteTargets  []RemoteTarget
	ActiveTarget   string // 当前激活的远程目标名称
	PathMappings   []PathMapping
	IgnorePatterns []*regexp.Regexp
	uploadChan     chan string
	watcher        *fsnotify.Watcher
}

func NewClient(mode, baseDir string) *Client {
	return &Client{
		Mode:           mode,
		LocalDir:       baseDir,
		RemoteTargets:  []RemoteTarget{},
		PathMappings:   []PathMapping{},
		IgnorePatterns: []*regexp.Regexp{},
		uploadChan:     make(chan string, NumWorkers),
	}
}

// AddRemoteTarget 添加一个远程目标
func (c *Client) AddRemoteTarget(name, url, targetDir, token string) {
	c.RemoteTargets = append(c.RemoteTargets, RemoteTarget{
		Name:      name,
		URL:       url,
		TargetDir: targetDir,
		Token:     token,
	})

	// 如果是第一个添加的目标，默认设为激活状态
	if len(c.RemoteTargets) == 1 {
		c.ActiveTarget = name
	}

	log.Printf("Added remote target: %s -> %s", name, url)
}

// SetActiveTarget 设置当前激活的远程目标
func (c *Client) SetActiveTarget(name string) error {
	for _, target := range c.RemoteTargets {
		if target.Name == name {
			c.ActiveTarget = name
			log.Printf("Activated remote target: %s", name)
			return nil
		}
	}
	return fmt.Errorf("remote target not found: %s", name)
}

// GetActiveTarget 获取当前激活的远程目标
func (c *Client) GetActiveTarget() (*RemoteTarget, error) {
	for _, target := range c.RemoteTargets {
		if target.Name == c.ActiveTarget {
			return &target, nil
		}
	}
	return nil, fmt.Errorf("no active remote target set")
}

// ListRemoteTargets 列出所有可用的远程目标
func (c *Client) ListRemoteTargets() []RemoteTarget {
	return c.RemoteTargets
}

// AddPathMapping 添加一个路径映射，使用正则表达式
func (c *Client) AddPathMapping(source, target string) {
	// 编译正则表达式
	pattern, err := regexp.Compile(source)
	if err != nil {
		log.Printf("Invalid path mapping pattern %s: %v", source, err)
		return
	}

	c.PathMappings = append(c.PathMappings, PathMapping{
		SourcePattern: pattern,
		TargetPath:    target,
	})
	log.Printf("Added path mapping: %s -> %s", source, target)
}

// AddIgnorePattern 添加一个忽略模式，使用正则表达式
func (c *Client) AddIgnorePattern(pattern string) {
	// 如果模式不是以^开头和$结尾，添加这些锚点以确保完全匹配
	if !strings.HasPrefix(pattern, "^") {
		pattern = "^" + pattern
	}
	if !strings.HasSuffix(pattern, "$") {
		pattern = pattern + "$"
	}
	
	// 编译正则表达式
	re, err := regexp.Compile(pattern)
	if err != nil {
		log.Printf("Invalid ignore pattern %s: %v", pattern, err)
		return
	}

	c.IgnorePatterns = append(c.IgnorePatterns, re)
	log.Printf("Added ignore pattern: %s", pattern)
}

// ShouldIgnore 检查文件是否应该被忽略，使用正则表达式匹配
func (c *Client) ShouldIgnore(path string) bool {
	// 检查文件是否匹配忽略正则表达式
	for _, pattern := range c.IgnorePatterns {
		if pattern.MatchString(path) {
			log.Println("ignore pattern: ", pattern, path)
			return true
		}
	}

	return false
}

// MapPath 根据路径映射规则映射路径，使用正则表达式
func (c *Client) MapPath(path string) string {
	// 如果没有路径映射规则，直接返回原始路径
	if len(c.PathMappings) == 0 {
		return path
	}

	// 检查是否有匹配的路径映射
	for _, mapping := range c.PathMappings {
		if mapping.SourcePattern.MatchString(path) {
			// 使用正则表达式替换
			return mapping.SourcePattern.ReplaceAllString(path, mapping.TargetPath)
		}
	}

	// 没有匹配的映射规则，返回原始路径
	return path
}

func (c *Client) Start() {
	// 启动 worker
	for w := 1; w <= NumWorkers; w++ {
		go c.worker(w)
	}

	var err error
	c.watcher, err = fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
		return
	}
	defer c.watcher.Close()

	var wg sync.WaitGroup
	wg.Add(1)
	go c.watcherThread()()
	wg.Add(1)
	go c.initNewDir()
	wg.Wait()
}

func (c *Client) watcherThread() func() {
	return func() {
		err := c.watcher.Add(c.LocalDir)
		if err != nil {
			log.Fatal(err)
		}

		for {
			select {
			case event, ok := <-c.watcher.Events:
				if !ok {
					return
				}
				log.Println(event)

				// 使用ShouldIgnore方法判断是否应该忽略文件
				if c.ShouldIgnore(event.Name) {
					log.Println("Ignoring file:", event.Name)
					continue
				}

				fi, err := os.Stat(event.Name)

				if event.Op&fsnotify.Create == fsnotify.Create {
					log.Println("Detected new file or directory:", event.Name)

					if err == nil && fi.IsDir() {
						// 检查目录是否应该被忽略
						if c.ShouldIgnore(event.Name) {
							log.Println("Ignoring directory:", event.Name)
							continue
						}

						errDir := c.watcher.Add(event.Name)
						log.Println("Watching new dir" + event.Name)
						if errDir != nil {
							log.Println("Error adding directory to watcher:", event.Name, errDir)
						}
						continue
					}

					time.Sleep(2 * time.Second) // 确保文件已完全写入
					c.uploadChan <- event.Name
				} else if event.Op&fsnotify.Write == fsnotify.Write {
					log.Println("Detected file change:", event.Name)
					time.Sleep(2 * time.Second) // 确保文件已完全写入
					c.uploadChan <- event.Name
				} else if event.Op&fsnotify.Remove == fsnotify.Remove {
					log.Println("Detected file removal, but ignore:", event.Name)
				} else if event.Op&fsnotify.Rename == fsnotify.Rename {
					log.Println("Detected file rename, but ignore:", event.Name)
				}
			case err, ok := <-c.watcher.Errors:
				if !ok {
					return
				}
				log.Println("Error:", err)
			}
		}
	}
}

func (c *Client) initNewDir() {
	// 初始化时上传文件
	var filesToUpload []string

	// 遍历本地目录收集文件
	err := filepath.Walk(c.LocalDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 检查是否应该忽略
		if c.ShouldIgnore(path) {
			if info.IsDir() {
				log.Println("Skipping ignored directory:", path)
				return filepath.SkipDir
			}
			log.Println("Skipping ignored file:", path)
			return nil
		}

		// 处理文件和目录
		if info.IsDir() {
			// 添加目录到监视器
			if err = c.watcher.Add(path); err != nil {
				log.Println("Failed to add directory to watcher:", path, err)
			}
			log.Println("Successfully added directory to watcher:", path)
		} else if c.Mode == "all" {
			// 全量模式下收集所有文件
			filesToUpload = append(filesToUpload, path)
		}
		
		return nil
	})
	
	if err != nil {
		log.Fatal("Failed to traverse directory:", err)
	}

	// 根据模式处理要上传的文件
	switch c.Mode {
	case "all":
		log.Printf("Preparing to upload all files: %d files", len(filesToUpload))
	case "git":
		// 获取Git差异文件
		gitFiles, err := getGitDiffFiles(c.LocalDir)
		if err != nil {
			log.Fatal("Failed to get Git diff files:", err)
		}

		// 过滤掉应该被忽略的文件
		filesToUpload = make([]string, 0, len(gitFiles))
		for _, file := range gitFiles {
			if !c.ShouldIgnore(file) {
				filesToUpload = append(filesToUpload, file)
			} else {
				log.Println("Skipping ignored Git diff file:", file)
			}
		}
		log.Printf("Preparing to upload Git diff files: %d files", len(filesToUpload))
	default:
		log.Fatalf("Unknown mode: %s", c.Mode)
	}

	// 将文件发送到上传通道
	for _, file := range filesToUpload {
		c.uploadChan <- file
	}
}

func (c *Client) uploadFile(filename string, baseDir string) error {
	activeTarget, err := c.GetActiveTarget()
	if err != nil {
		return err
	}

	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)
	err = writer.WriteField("token", activeTarget.Token)
	if err != nil {
		return err
	}
	part, err := writer.CreateFormFile("file", filepath.Base(filename))
	if err != nil {
		return err
	}
	if _, err := io.Copy(part, file); err != nil {
		return err
	}

	relativePath, err := filepath.Rel(baseDir, filename)
	if err != nil {
		return err
	}

	// 应用路径映射
	mappedPath := c.MapPath(filepath.Join(activeTarget.TargetDir, relativePath))
	targetPath := mappedPath

	log.Printf("Uploading file to %s: %s", activeTarget.Name, targetPath)

	writer.WriteField("target", targetPath)
	contentType := writer.FormDataContentType()
	writer.Close()

	// 创建带有超时时间的客户端
	client := &http.Client{
		Timeout: 30 * time.Second, // 设置超时
	}
	
	// 创建请求
	req, err := http.NewRequest("POST", activeTarget.URL, &requestBody)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", contentType)
	
	// 发送请求
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	// 读取 body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Failed to read response body: %v", err)
		return fmt.Errorf("failed to read response body: %w", err)
	}

	log.Printf("upload res: %+v", string(body))

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to upload file: %s", resp.Status)
	}

	return nil
}

func (c *Client) worker(id int) {
	for filename := range c.uploadChan {
		err := c.uploadFile(filename, c.LocalDir)
		if err != nil {
			log.Printf("Worker %d failed to upload file: %s error: %v\n", id, filename, err)
		}
	}
}

func getGitDiffFiles(baseDir string) ([]string, error) {
	cmd := exec.Command("git", "diff", "origin/master", "--name-only")
	cmd.Dir = baseDir
	output, err := cmd.Output()
	if err != nil {
		log.Printf("Failed to get git diff files: %v try origin/main", err)
		cmd := exec.Command("git", "diff", "origin/main", "--name-only")
		cmd.Dir = baseDir
		output, err = cmd.Output()
		if err != nil {
			log.Printf("Failed to get git diff files: %v", err)
			return nil, err
		}
	}

	files := strings.Split(string(output), "\n")
	var diffFiles []string
	for _, file := range files {
		if file != "" {
			diffFiles = append(diffFiles, filepath.Join(baseDir, file))
		}
	}
	return diffFiles, nil
}
