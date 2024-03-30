package main

import (
	"bufio"
	"bytes"
	"encoding/gob"
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/winfsp/cgofuse/fuse"
)

const (
	filename = "hello world"
	contents = "hello, world\n"
	file     = "hello amanaaaaaaaa"
)

type Hellofs struct {
	fuse.FileSystemBase
}

type ReaddirS struct {
	Path string
	Args map[string]interface{}
}

func (*Hellofs) Open(path string, flags int) (errc int, fh uint64) {
	switch path {
	case "/" + filename:
		return 0, 0
	default:
		return -fuse.ENOENT, ^uint64(0)
	}
}

func (*Hellofs) Getattr(path string, stat *fuse.Stat_t, fh uint64) (errc int) {
	switch path {
	case "/":
		stat.Mode = fuse.S_IFDIR | 0555
		return 0
	case "/" + filename:
		stat.Mode = fuse.S_IFREG | 0444
		stat.Size = int64(len(contents))
		return 0

	default:
		stat.Mode = fuse.S_IFREG | 0444
		stat.Size = int64(len(contents))
		return 0
	}
}

func (*Hellofs) Read(path string, buff []byte, ofst int64, fh uint64) (n int) {
	endofst := ofst + int64(len(buff))
	if endofst > int64(len(contents)) {
		endofst = int64(len(contents))
	}
	if endofst < ofst {
		return 0
	}
	n = copy(buff, contents[ofst:endofst])
	return
}

func (*Hellofs) Readdir(path string,
	fill func(name string, stat *fuse.Stat_t, ofst int64) bool,
	ofst int64,
	fh uint64) (errc int) {

	fill(".", nil, 0)
	fill("..", nil, 0)

	f, err := os.Open("./")
	if err != nil {
		fmt.Println(err)
		return
	}
	files, err := f.Readdir(0)
	if err != nil {
		fmt.Println(err)
		return
	}

	for _, v := range files {
		fill(v.Name(), nil, 0)
	}

	return 0
}

func main() {

	gob.Register(map[string]interface{}{})

	var s = ReaddirS{"readdir", map[string]any{
		"path": "./",
		"ofst": 6,
		"fh":   0,
	}}

	b := bytes.Buffer{}
	e := gob.NewEncoder(&b)

	err := e.Encode(s)
	if err != nil {
		fmt.Println(`failed gob Encode`, err)
	}

	//fmt.Println(b.Bytes())

	// m := ReaddirS{}
	// c := bytes.Buffer{}
	// c.Write(b.Bytes())
	// d := gob.NewDecoder(&c)
	// err = d.Decode(&m)
	// if err != nil {
	// 	fmt.Println(`failed gob Decode`, err)
	// }

	// fmt.Println(m.Args["ofst"])

	// hellofs := &Hellofs{}
	// host := fuse.NewFileSystemHost(hellofs)
	// host.Mount("", os.Args[1:])

	c, err := net.Dial("tcp", "localhost:3000")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	for {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print(">> ")
		text, _ := reader.ReadString('\n')
		fmt.Fprintf(c, text+"\n")

		message, _ := bufio.NewReader(c).ReadString('\n')
		fmt.Print("->: " + message)
		if strings.TrimSpace(string(text)) == "STOP" {
			c.Close()
			fmt.Println("TCP client exiting...")
			return
		}
	}
}
