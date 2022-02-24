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
	var length = "Content-Length: " + strconv.FormatInt(int64(len(content)), 10)
	var response []byte = []byte(status)
	response = append(response, []byte(cr)...)
	response = append(response, []byte(mimetype)...)
	response = append(response, []byte(cr)...)
	response = append(response, []byte(noSniff)...)
	response = append(response, []byte(cr)...)
	response = append(response, []byte(length)...)
	response = append(response, []byte(cr)...)
	response = append(response, []byte(cr)...)
	response = append(response, content...)
	response = append(response, []byte(cr)...)
	response = append(response, []byte(cr)...)
	return response
}
