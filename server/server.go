package server

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
)

type Server struct {
	Port     int
	Token    string
	LimitDir string
}

func NewServer(port int, token, limitDir string) *Server {
	return &Server{
		Port:     port,
		Token:    token,
		LimitDir: limitDir,
	}
}

func (s *Server) Start() {
	// 捕获 ctrl-c 信号，并关闭服务器
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT)

	go func() {
		<-sigs
		log.Println("Caught SIGINT, stopping server...")
		os.Exit(0)
	}()

	http.HandleFunc("/receiver", s.uploadHandler)
	log.Printf("Starting server at port %d...\n", s.Port)
	log.Println("Limit directory: ", s.LimitDir)
	log.Println("Token: ", s.Token)

	if err := http.ListenAndServe(fmt.Sprintf(":%d", s.Port), nil); err != nil {
		log.Fatal(err)
	}
}

func (s *Server) uploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	if len(s.Token) != 0 {
		uToken := r.PostFormValue("token")
		if uToken != s.Token {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	fullPath := r.FormValue("target")
	if fullPath == "" {
		http.Error(w, "Missing target", http.StatusBadRequest)
		return
	}
	log.Println("Uploading to: ", fullPath)

	if !filepath.IsAbs(fullPath) || !strings.HasPrefix(fullPath, s.LimitDir) {
		http.Error(w, "Invalid target path, valid path: "+s.LimitDir, http.StatusBadRequest)
		return
	}

	// 创建目录
	// 判断目录是否存在
	if _, err := os.Stat(filepath.Dir(fullPath)); os.IsNotExist(err) {
		if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Println(err)
			return
		}
	}

	var out *os.File
	// 创建文件
	out, err = os.Create(fullPath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println(err)
		return
	}
	defer out.Close()

	// 将上传的文件内容写入到新文件中
	if _, err := io.Copy(out, file); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println(err)
		return
	}

	fmt.Fprintf(w, "File uploaded successfully: %s\n", fullPath)
}
