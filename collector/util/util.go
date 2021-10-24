package util

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
)

func PathExists(path string) (bool) {
	_, err := os.Stat(path)
	if err != nil &&  os.IsNotExist(err){
		return true
	}
	return false
}

func Mkdir(dir string)  {
	if PathExists(dir) {
		os.Mkdir(dir, os.ModePerm)
	}
}

func JsonFormatting(reader io.Reader) io.Reader {
	data, err := io.ReadAll(reader)
	if err != nil {
		panic(err)
	}
	buffer := bytes.Buffer{}
	err = json.Indent(&buffer, data, "", " ")
	if err != nil {
		panic(err)
	}
	return &buffer
}