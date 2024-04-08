package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/winfsp/cgofuse/fuse"
)

const (
	url = "http://localhost:3000/"
)

type Hellofs struct {
	fuse.FileSystemBase
}

type Operation struct {
	Op   string
	Args map[string]interface{}
}

func (*Hellofs) Open(path string, flags int) (errc int, fh uint64) {
	return 0, 0
}

func (*Hellofs) Getattr(path string, stat *fuse.Stat_t, fh uint64) (errc int) {
	switch path {
	case "/":
		stat.Mode = fuse.S_IFDIR | 0555
		return 0

	default:
		data := PerformOperation("getattr", path, map[string]interface{}{})

		var m = struct {
			IsDir bool
			Size  int64
		}{}

		c := bytes.Buffer{}
		c.Write(data)
		d := gob.NewDecoder(&c)
		err := d.Decode(&m)
		if err != nil {
			fmt.Println(`failed gob Decode`, err)
		}

		stat.Size = m.Size

		if m.IsDir {
			stat.Mode = fuse.S_IFDIR | 0555
		} else {
			stat.Mode = fuse.S_IFREG | 0444
		}

		return 0
	}
}

func (*Hellofs) Read(path string, buff []byte, ofst int64, fh uint64) (n int) {

	endofst := ofst + int64(len(buff))

	if endofst < ofst {
		return 0
	}

	data := PerformOperation("read", path, map[string]interface{}{
		"offset": ofst,
		"size":   endofst - ofst,
	})

	n = copy(buff, data)
	return
}

func (*Hellofs) Readdir(path string,
	fill func(name string, stat *fuse.Stat_t, ofst int64) bool,
	ofst int64,
	fh uint64) (errc int) {

	fill(".", nil, 0)
	fill("..", nil, 0)

	// file, err := os.Open("./" + path)
	// if err != nil {
	// 	fmt.Printf("failed opening directory: %s", err)
	// }
	// defer file.Close()

	// fileList, _ := file.Readdir(0)

	// for _, file := range fileList {
	// 	fill(file.Name(), nil, 0)
	// }

	data := PerformOperation("readdir", path, map[string]interface{}{})

	var m = []struct {
		Name  string
		IsDir bool
		Size  int64
	}{}
	c := bytes.Buffer{}
	c.Write(data)
	d := gob.NewDecoder(&c)
	err := d.Decode(&m)
	if err != nil {
		fmt.Println(`failed gob Decode`, err)
	}

	for _, file := range m {
		stat := fuse.Stat_t{}
		stat.Size = file.Size
		if file.IsDir {
			stat.Mode = fuse.S_IFDIR | 0555
		} else {
			stat.Mode = fuse.S_IFREG | 0444
		}
		fill(file.Name, &stat, 0)
	}

	return 0
}

func (*Hellofs) Write(path string, buff []byte, ofst int64, fh uint64) (n int) {

	PerformOperation("write", path, map[string]interface{}{
		"offset": ofst,
		"data":   buff,
	})

	return
}

func (*Hellofs) Statfs(path string, stat *fuse.Statfs_t) (errc int) {

	stat.Bavail = 1000000
	stat.Bsize = 4096
	stat.Blocks = 1000000
	stat.Bfree = 1000000
	stat.Frsize = 4096
	stat.Files = 1000000
	stat.Ffree = 1000000
	stat.Favail = 1000000
	stat.Namemax = 255

	return
}

func PerformOperation(op string, path string, args map[string]interface{}) []byte {

	var s Operation
	switch op {

	case "readdir":
		s = Operation{"readdir", map[string]any{
			"path": path,
			"args": args,
		}}

	case "getattr":
		s = Operation{"getattr", map[string]any{
			"path": path,
			"args": args,
		}}

	case "read":
		s = Operation{"read", map[string]any{
			"path": path,
			"args": args,
		}}

	case "write":
		s = Operation{"write", map[string]any{
			"path": path,
			"args": args,
		}}

	}

	b := bytes.Buffer{}
	e := gob.NewEncoder(&b)

	err := e.Encode(s)
	if err != nil {
		fmt.Println(`failed gob Encode`, err)
	}

	resp, err := http.Post(url, "application/octet-stream", &b)

	if err != nil {
		fmt.Println(`failed http post`, err)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(`failed to read response body`, err)
	}

	resp.Body.Close()

	return bodyBytes

}

func main() {
	gob.Register(map[string]interface{}{})

	hellofs := &Hellofs{}
	host := fuse.NewFileSystemHost(hellofs)
	host.Mount("", os.Args[1:])
}
