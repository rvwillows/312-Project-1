package main

import (
	"bytes"
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

func parseRequest(buffer []byte) {
	splitBuffer := bytes.SplitN(buffer, []byte("\r\n\r\n"), 2)
	var headers = string(splitBuffer[0])
	var body = splitBuffer[1]

	var mimetype = strings.Split(strings.Split(headers, "Content-Type: ")[1], ";")[0]

	if mimetype == "multipart/form-data" {
		comment := Comment{}
		var boundary = strings.Split(strings.Split(headers, "boundary=")[1], "\r\n")[0]
		content := bytes.Split(body, []byte("--"+boundary))
		for _, c := range content {
			if bytes.Contains(c, []byte("\r\n\r\n")) {
				subBytes := bytes.Split(c, []byte("\r\n\r\n"))
				var subHeader = string(subBytes[0])
				var subContent = subBytes[1]
				if strings.Contains(subHeader, `name="comment"`) {
					comment.Message = string(subContent)
				} else if strings.Contains(subHeader, `name="upload"`) {
					comment.Image = strings.Split(strings.Split(subHeader, "filename=")[1], "\r\n")[0]
				}
			}
		}
		addComment(comment)
	}
}
