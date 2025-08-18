package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	logger "sockect-demo/logger"
	"strings"
	"sync"
)

const ROOTDIR = "data"

var (
	clients   = make(map[string]net.Conn) // 记录哪些客户端连接
	clientMux sync.Mutex
)

func main() {
	logger.SetLevel(logger.DEBUG)

	address := "127.0.0.1:2121"

	l, err := net.Listen("tcp", address)
	if err != nil {
		logger.Errorf("Listen on %s error: %s", address, err)
		return
	}

	defer l.Close()
	logger.Info("Listen on", address)

	for {
		conn, err := l.Accept()

		clientMux.Lock()
		clients[conn.RemoteAddr().String()] = conn
		clientMux.Unlock()

		if err != nil {
			logger.Errorf("Connect on %s error: %s", conn.RemoteAddr(), err)
			continue
		}

		go fileServer(conn)
	}
}

func fileServer(conn net.Conn) {
	defer func() {
		defer conn.Close()
		clientMux.Lock()
		delete(clients, conn.RemoteAddr().String())
		clientMux.Unlock()

		logger.Infof("Connect %s closed.", conn.RemoteAddr())
	}()

	scanner := bufio.NewScanner(conn)

	for scanner.Scan() {
		var data string
		var err error
		text := scanner.Text()
		logger.Debug("Get text:", text)
		cmd := strings.ToLower(text)

		if cmd == "list" {
			logger.Debug("Should run listServe")
			data, err = PrintTree(ROOTDIR, "")
			if err != nil {
				logger.Errorf("list dir %s error: %s", ROOTDIR, err)
				continue
			}

			logger.Debugf("Data: \n %s", data)

		} else if strings.HasPrefix(cmd, "delete") {
			logger.Debug("Should run delete serve")
			data, err = deleteServe(cmd)
			if err != nil {
				logger.Errorf("delete error: %s", err)
				continue
			}

			logger.Debugf("Data: \n %s", data)
		} else if strings.HasPrefix(cmd, "upload") {
			logger.Debug("Should run upload serve")
			data, err = uploadServe(cmd)
			if err != nil {
				logger.Errorf("delete error: %s", err)
				continue
			}
		}

		if _, err = conn.Write([]byte(data)); err != nil {
			logger.Errorf("Write data to %s error: %s \n", conn.RemoteAddr(), err)
		}
	}
}

func PrintTree(root string, prefix string) (string, error) {
	var result strings.Builder

	entries, err := os.ReadDir(root)
	if err != nil {
		return "", err
	}

	for i, entry := range entries {
		connector := "├── "
		newPrefix := prefix + "│   "
		if i == len(entries)-1 {
			connector = "└── "
			newPrefix = prefix + "    "
		}

		result.WriteString(prefix + connector + entry.Name() + "\n")

		if entry.IsDir() {
			subTree, err := PrintTree(filepath.Join(root, entry.Name()), newPrefix)
			if err != nil {
				return "", err
			}
			result.WriteString(subTree)
		}
	}
	return result.String(), nil
}

func deleteServe(cmd string) (string, error) {
	data := ""
	parts := strings.Split(cmd, " ")
	logger.Debug("cmd parts: ", parts)

	if len(parts) <= 1 {
		data = "must provide delete file path \n"
		return data, nil
	}

	paths := parts[1:]
	for _, path := range paths {
		path = filepath.Join(ROOTDIR, path)
		logger.Debug("Path: ", path)

		if err := os.Remove(path); err != nil {
			msg := fmt.Sprintf("delete %s error: %s", path, err)
			data += msg + " "
			logger.Error(msg)
			continue
		}

		msg := fmt.Sprintf("delete %s success", path)
		data += msg + " "
		logger.Info(msg)
	}

	return data + "\n", nil
}

// upload file1 file2 means move file1 to file2
func uploadServe(cmd string) (string, error) {
	parts := strings.Split(cmd, " ")
	logger.Debug("cmd parts: ", parts)

	if len(parts) < 3 {
		return "usage: upload src dst\n", nil
	}

	src := parts[1]
	dst := filepath.Join(ROOTDIR, parts[2])
	logger.Debugf("Copy file %s → %s", src, dst)

	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return "", fmt.Errorf("create dir error: %w", err)
	}

	// 打开源文件
	in, err := os.Open(src)
	if err != nil {
		return "", fmt.Errorf("open src file error: %w", err)
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return "", fmt.Errorf("create dst file error: %w", err)
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return "", fmt.Errorf("copy file error: %w", err)
	}

	if err := out.Sync(); err != nil {
		return "", fmt.Errorf("sync file error: %w", err)
	}

	return fmt.Sprintf("upload (copy) %s → %s success\n", src, dst), nil
}
