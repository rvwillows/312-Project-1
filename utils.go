package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

func loadFile(path string) []byte {
	if strings.HasPrefix(path, "/") {
		path = "." + path
	} else {
		path = "./" + path
	}
	fmt.Println(path)
	dat, err := os.ReadFile(path)
	if err == nil {
		return dat
	} else {
		fmt.Println("File read error: ", err.Error())
		return nil
	}
}

func makeResponse(status string, mimetype string, content []byte) []byte {
	var length = 0
	if status != moved {
		length = len(content)
	}
	var response []byte = []byte(status)
	response = append(response, []byte(cr)...)
	if status != moved {
		response = append(response, []byte(mimetype)...)
		response = append(response, []byte(cr)...)
		response = append(response, []byte(noSniff)...)
		response = append(response, []byte(cr)...)
	}
	response = append(response, []byte("Content-Length: "+strconv.FormatInt(int64(length), 10))...)
	response = append(response, []byte(cr)...)
	if status != moved {
		response = append(response, []byte(cr)...)
	}
	response = append(response, content...)
	response = append(response, []byte(cr)...)
	response = append(response, []byte(cr)...)
	return response
}
