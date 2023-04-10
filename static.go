package main

import (
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
)

// 内嵌资源目录指令
//
//go:embed *.conf
var conf embed.FS

//go:embed index.html
var index embed.FS

type embedFileSystem struct {
	http.FileSystem
}

func (e embedFileSystem) Exists(prefix, path string) bool {
	_, err := e.Open(path)
	if err != nil {
		return false
	}
	return true
}

func getFileSystem(useOS bool) http.FileSystem {
	if useOS {
		log.Print("using live mode")
		return http.FS(os.DirFS("static"))
	}
	fsys, err := fs.Sub(conf, "static")
	if err != nil {
		panic(err)
	}
	return http.FS(fsys)
}

// ReleaseFile 释放文件到指定路径
func ReleaseFile(fsEmbed embed.FS, targetPath string) {
	/*fsys, err := fs.Sub(fsEmbed, targetPath)
	  if err != nil {
	  	panic(err)
	  }*/
	content, err := fsEmbed.ReadFile(targetPath)
	if err != nil {
		log.Fatal(err)
	}
	// 创建文件
	file, err := os.OpenFile(targetPath, os.O_WRONLY|os.O_CREATE, fs.FileMode.Perm(0666)) //0666 在windows下无效
	if err != nil {
		fmt.Println("open file err:", err)
		return
	}
	//关闭文件
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(file)
	//写入
	err = os.WriteFile(targetPath, content, fs.FileMode.Perm(0666))
	if err != nil {
		log.Fatal(err)
	}
}
