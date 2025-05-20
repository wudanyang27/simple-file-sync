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
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

const (
	NumWorkers = 30
)

type PathMapping struct {
	SourcePath string
	TargetPath string
}

type Client struct {
	Mode            string
	LocalDir        string
	ServerURL       string
	ServerTargetDir string
	ServerToken     string
	PathMappings    []PathMapping
	IgnorePatterns  []string
	uploadChan      chan string
	watcher         *fsnotify.Watcher
}

func NewClient(mode, baseDir, serverURL, target, token string) *Client {
	return &Client{
		Mode:            mode,
		LocalDir:        baseDir,
		ServerURL:       serverURL,
		ServerTargetDir: target,
		ServerToken:     token,
		PathMappings:    []PathMapping{},
		IgnorePatterns:  []string{},
		uploadChan:      make(chan string, NumWorkers),
	}
}

// AddPathMapping 添加一个路径映射
func (c *Client) AddPathMapping(source, target string) {
	c.PathMappings = append(c.PathMappings, PathMapping{
		SourcePath: source,
		TargetPath: target,
	})
}

// AddIgnorePattern 添加一个忽略模式
func (c *Client) AddIgnorePattern(pattern string) {
	c.IgnorePatterns = append(c.IgnorePatterns, pattern)
}

// ShouldIgnore 检查文件是否应该被忽略
func (c *Client) ShouldIgnore(path string) bool {
	// 忽略 .DS_Store 文件
	if strings.HasSuffix(path, ".DS_Store") {
		return true
	}

	// 忽略备份文件
	if strings.HasSuffix(path, "~") {
		return true
	}

	// 检查文件是否匹配忽略模式
	for _, pattern := range c.IgnorePatterns {
		matched, err := filepath.Match(pattern, filepath.Base(path))
		if err == nil && matched {
			return true
		}

		// 检查路径是否匹配模式
		if strings.Contains(path, pattern) {
			return true
		}
	}

	// 检查是否是隐藏目录
	fi, err := os.Stat(path)
	if err == nil && fi.IsDir() && strings.HasPrefix(fi.Name(), ".") {
		return true
	}

	return false
}

// MapPath 根据路径映射规则映射路径
func (c *Client) MapPath(path string) string {
	// 如果没有路径映射规则，直接返回原始路径
	if len(c.PathMappings) == 0 {
		return path
	}

	// 检查是否有匹配的路径映射
	for _, mapping := range c.PathMappings {
		if strings.HasPrefix(path, mapping.SourcePath) {
			// 替换前缀
			return strings.Replace(path, mapping.SourcePath, mapping.TargetPath, 1)
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
				if err == nil && fi.IsDir() && strings.HasPrefix(fi.Name(), ".") {
					log.Println("Skipping hidden directory:", event.Name)
					continue
				}

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

	err := filepath.Walk(c.LocalDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 使用ShouldIgnore方法判断是否应该忽略文件或目录
		if c.ShouldIgnore(path) {
			if info.IsDir() {
				log.Println("Skipping ignored directory:", path)
				return filepath.SkipDir
			}
			log.Println("Skipping ignored file:", path)
			return nil
		}

		if !info.IsDir() {
			if c.Mode == "all" {
				filesToUpload = append(filesToUpload, path)
			}
		} else {
			err = c.watcher.Add(path)
			if err != nil {
				log.Println("Error adding directory to watcher:", path, err)
			}
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}

	if c.Mode == "all" {
		log.Println("Uploading all files:", len(filesToUpload))
	} else if c.Mode == "git" {
		filesToUpload, err = getGitDiffFiles(c.LocalDir)
		if err != nil {
			log.Fatal(err)
		}

		// 过滤掉应该被忽略的文件
		var filteredFiles []string
		for _, file := range filesToUpload {
			if !c.ShouldIgnore(file) {
				filteredFiles = append(filteredFiles, file)
			} else {
				log.Println("Skipping ignored git diff file:", file)
			}
		}
		filesToUpload = filteredFiles

		log.Println("Uploading git diff files:", len(filesToUpload))
	} else {
		log.Fatalf("Unknown mode: %s", c.Mode)
	}

	for _, file := range filesToUpload {
		c.uploadChan <- file
	}
}

func (c *Client) uploadFile(filename string, serverURL string, target string, baseDir string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)
	err = writer.WriteField("token", c.ServerToken)
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
	mappedPath := c.MapPath(filepath.Join(target, relativePath))
	targetPath := mappedPath

	log.Println("Uploading file: ", targetPath)

	writer.WriteField("target", targetPath)
	contentType := writer.FormDataContentType()
	writer.Close()

	resp, err := http.Post(serverURL, contentType, &requestBody)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	// 读取 body
	body, _ := io.ReadAll(resp.Body)

	log.Printf("upload res: %+v", string(body))

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to upload file: %s", resp.Status)
	}

	return nil
}

func (c *Client) worker(id int) {
	for filename := range c.uploadChan {
		err := c.uploadFile(filename, c.ServerURL, c.ServerTargetDir, c.LocalDir)
		if err != nil {
			log.Printf("Worker %d failed to upload file: %s error: %v\n", id, filename, err)
		}
	}
}

func getGitDiffFiles(baseDir string) ([]string, error) {
	cmd := exec.Command("git", "diff", "origin", "--name-only")
	cmd.Dir = baseDir
	output, err := cmd.Output()
	if err != nil {
		return nil, err
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
