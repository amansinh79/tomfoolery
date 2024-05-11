package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

var (
	rootPath = os.Args[1]
)

func root(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusInternalServerError)
		return
	}

	m := struct {
		Op   string
		Args map[string]interface{}
	}{}

	c := bytes.Buffer{}
	c.Write(body)
	d := gob.NewDecoder(&c)
	err = d.Decode(&m)

	if err != nil {
		fmt.Println(`failed gob Decode`, err)
	}

	switch m.Op {
	case "readdir":
		path := m.Args["path"].(string)

		file, err := os.Open(rootPath + path)
		if err != nil {
			log.Fatalf("failed opening directory: %s", err)
		}
		defer file.Close()

		fileList, _ := file.Readdir(0)

		w.Header().Set("Content-Type", "application/octet-stream")

		b := bytes.Buffer{}
		e := gob.NewEncoder(&b)

		files := []struct {
			Name  string
			IsDir bool
			Size  int64
		}{}

		for _, f := range fileList {
			files = append(files, struct {
				Name  string
				IsDir bool
				Size  int64
			}{
				Name:  f.Name(),
				IsDir: f.IsDir(),
				Size:  f.Size(),
			})
		}

		err = e.Encode(files)

		if err != nil {
			fmt.Println(`failed gob Encode`, err)
		}

		w.Write(b.Bytes())

	case "getattr":
		path := m.Args["path"].(string)

		var m = struct {
			IsDir bool
			Size  int64
		}{}
		file, err := os.Open(rootPath + path)
		if err == nil {
			defer file.Close()
			s, err := file.Stat()
			if err == nil {
				m.Size = s.Size()
				m.IsDir = s.IsDir()
			}
		}

		b := bytes.Buffer{}
		e := gob.NewEncoder(&b)

		err = e.Encode(m)

		if err != nil {
			fmt.Println(`failed gob Encode`, err)
		}

		w.Write(b.Bytes())

	case "read":
		path := m.Args["path"].(string)

		args := m.Args["args"]

		argsMap := args.(map[string]interface{})

		ofst := argsMap["offset"].(int64)
		size := argsMap["size"].(int64)

		_, err := os.Stat(rootPath + path)

		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			w.Write(make([]byte, 0))
			return
		}

		file, err := os.Open(rootPath + path)

		if err != nil {
			log.Fatalf("failed opening file: %s", err)
		}
		defer file.Close()

		_, err = file.Seek(ofst, io.SeekStart)
		if err != nil {

			log.Fatalf("failed seeking file: %s", err)
		}

		b := make([]byte, size)

		_, err = file.Read(b)
		if err != nil {
			log.Fatalf("failed reading file: %s", err)
		}

		w.Header().Set("Content-Type", "application/octet-stream")

		w.Write(b)

	case "open":
		{
			path := m.Args["path"].(string)
			file, err := os.Open(rootPath + path)

			var m = struct {
				Fh uint64
			}{}

			if err == nil {
				m.Fh = uint64(file.Fd())
				file.Close()
			} else {
				m.Fh = 0
			}

			b := bytes.Buffer{}

			e := gob.NewEncoder(&b)

			err = e.Encode(m)

			if err != nil {
				fmt.Println(`failed gob Encode`, err)
			}
			w.Write(b.Bytes())
		}
	}
}

func main1() {
	gob.Register(map[string]interface{}{})
	fmt.Println(rootPath)
	http.HandleFunc("/", root)
	http.ListenAndServe(":3000", nil)
}
