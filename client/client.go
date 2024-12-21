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

type Client struct {
	Mode            string
	LocalDir        string
	ServerURL       string
	ServerTargetDir string
	ServerToken     string
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
		uploadChan:      make(chan string, NumWorkers),
	}
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
				if strings.HasSuffix(event.Name, ".DS_Store") {
					continue
				}
				if strings.HasSuffix(event.Name, "~") {
					log.Println("Skipping backup file:", event.Name)
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
		if !info.IsDir() {
			if c.Mode == "all" {
				filesToUpload = append(filesToUpload, path)
			}
		} else {
			if strings.HasPrefix(info.Name(), ".") {
				log.Println("Skipping hidden directory:", path)
				return filepath.SkipDir
			}
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

	targetPath := filepath.Join(target, relativePath)
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
	body, err := io.ReadAll(resp.Body)

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
