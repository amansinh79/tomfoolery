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
	url = "http://localhost:4000/"
)

type Fs struct {
	fuse.FileSystemBase
}

type Operation struct {
	Op   string
	Args map[string]interface{}
}

func (*Fs) Open(path string, flags int) (errc int, fh uint64) {

	data := PerformOperation("open", path, map[string]interface{}{
		"flags": flags,
	})

	var m = struct {
		Fh uint64
	}{}

	c := bytes.Buffer{}
	c.Write(data)

	d := gob.NewDecoder(&c)
	err := d.Decode(&m)
	if err != nil {
		fmt.Println(`failed gob Decode`, err)
	}

	return 0, m.Fh
}

func (*Fs) Getattr(path string, stat *fuse.Stat_t, fh uint64) (errc int) {
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

func (*Fs) Read(path string, buff []byte, ofst int64, fh uint64) (n int) {

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

func (*Fs) Readdir(path string,
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

func PerformOperation(op string, path string, args map[string]interface{}) []byte {

	var s = Operation{op, map[string]any{
		"path": path,
		"args": args,
	}}

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

	hellofs := &Fs{}
	host := fuse.NewFileSystemHost(hellofs)
	host.Mount("", os.Args[1:])
}
