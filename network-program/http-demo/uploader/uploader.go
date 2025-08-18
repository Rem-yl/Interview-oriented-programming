// go run uploader.go
// 在浏览器中访问 http://127.0.0.1:8088, 在文件框中输入文件路径
package main

import (
	"fmt"
	"http-demo/logger"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

const (
	ROOTPATH = "./data"
)

// 初始化函数，确保上传目录存在
func init() {
	if err := os.MkdirAll(ROOTPATH, 0755); err != nil {
		logger.Errorf("无法创建上传目录: %v\n", err)
		os.Exit(1)
	}
}

func uploadPage(w http.ResponseWriter, r *http.Request) {
	// 设置正确的Content-Type
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	html := `
	<html>
	<head>
		<title>File Upload by Path</title>
		<style>
			body { font-family: Arial, sans-serif; max-width: 800px; margin: 2rem auto; padding: 0 1rem; }
			.form-group { margin: 1rem 0; }
			input[type="text"] { 
				width: 100%; 
				padding: 0.5rem; 
				box-sizing: border-box;
			}
			input[type="submit"] { 
				background-color: #4CAF50; 
				color: white; 
				padding: 0.7rem 1.2rem; 
				border: none; 
				border-radius: 4px; 
				cursor: pointer;
			}
			input[type="submit"]:hover { background-color: #45a049; }
			.message { margin-top: 1rem; padding: 1rem; border-radius: 4px; }
			.success { background-color: #dff0d8; color: #3c763d; }
			.error { background-color: #f2dede; color: #a94442; }
		</style>
	</head>
	<body>
		<h1>文件上传服务</h1>
		<form action="/upload" method="post">
			<div class="form-group">
				<label for="filePath">请输入文件路径:</label><br>
				<input type="text" id="filePath" name="filePath" placeholder="例如: /home/user/documents/file.txt 或 C:\files\data.pdf" required>
			</div>
			<div class="form-group">
				<input type="submit" value="上传文件">
			</div>
		</form>
	</body>
	</html>`
	_, err := w.Write([]byte(html))
	if err != nil {
		logger.Errorf("写入响应失败: %v\n", err)
	}
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "只允许POST方法", http.StatusMethodNotAllowed)
		return
	}

	// 获取用户输入的文件路径
	filePath := r.FormValue("filePath")
	if filePath == "" {
		http.Error(w, "文件路径不能为空", http.StatusBadRequest)
		return
	}

	// 检查文件是否存在
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		http.Error(w, "文件不存在或无法访问: "+err.Error(), http.StatusBadRequest)
		return
	}

	// 检查是否是目录
	if fileInfo.IsDir() {
		http.Error(w, "指定的路径是一个目录，不是文件", http.StatusBadRequest)
		return
	}

	// 打开源文件
	srcFile, err := os.Open(filePath)
	if err != nil {
		http.Error(w, "无法打开文件: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer srcFile.Close()

	// 获取文件名
	filename := filepath.Base(filePath)
	savePath := filepath.Join(ROOTPATH, filename)

	// 检查文件是否已存在
	if _, err := os.Stat(savePath); err == nil {
		http.Error(w, "目标文件已存在", http.StatusConflict)
		return
	}

	// 创建目标文件
	dstFile, err := os.Create(savePath)
	if err != nil {
		http.Error(w, "无法创建目标文件: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer dstFile.Close()

	// 复制文件内容
	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		http.Error(w, "复制文件内容失败: "+err.Error(), http.StatusInternalServerError)
		// 尝试删除已创建的空文件
		os.Remove(savePath)
		return
	}

	// 输出成功信息
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, `
	<html>
	<head>
		<title>上传成功</title>
		<style>
			/* 复用上传页面的样式 */
			body { font-family: Arial, sans-serif; max-width: 800px; margin: 2rem auto; padding: 0 1rem; }
			.success { background-color: #dff0d8; color: #3c763d; padding: 1rem; border-radius: 4px; }
			a { color: #337ab7; text-decoration: none; }
			a:hover { text-decoration: underline; }
		</style>
	</head>
	<body>
		<div class="success">
			<h2>上传成功!</h2>
			<p>源文件路径: %s</p>
			<p>文件大小: %d 字节</p>
			<p>保存路径: %s</p>
		</div>
		<p style="margin-top: 1rem;"><a href="/">返回上传页面</a></p>
	</body>
	</html>`, filePath, fileInfo.Size(), savePath)
}

func main() {
	address := "127.0.0.1:8088"

	http.HandleFunc("/", uploadPage)
	http.HandleFunc("/upload", uploadHandler)

	fmt.Printf("服务器启动，监听地址: %s\n", address)
	if err := http.ListenAndServe(address, nil); err != nil {
		fmt.Printf("服务器启动失败: %s\n", err)
	}
}
