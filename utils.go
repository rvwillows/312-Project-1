package main

import (
	"bytes"
	"fmt"
	"net"
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
		if strings.Contains(path, "index.html") {
			comments := getComments()
			var commentString = ""
			for _, comment := range comments {
				commentString = commentString + "<br>" + comment.Image + "</br>"
				commentString = commentString + "<br> <img src=" + comment.Image + "> </br>"
				commentString = commentString + "<br>" + comment.Message + "</br>"
			}
			dat = []byte(strings.Replace(string(dat), "{{data}}", commentString, 1))
		}
		return dat
	} else {
		fmt.Println("File read error: ", err.Error())
		return nil
	}
}

func saveFile(path string, data []byte) {
	file, err := os.Create(path)
	if err != nil {
		fmt.Println("File write error: ", err.Error())
	}
	defer file.Close()

	n, err := file.Write(data)
	if err != nil {
		fmt.Println("File write error: ", err.Error())
	}
	fmt.Println("wrote %d bytes to "+path, n)

	file.Sync()
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

func parseRequest(buffer []byte, conn net.Conn) {
	splitBuffer := bytes.SplitN(buffer, []byte("\r\n\r\n"), 2)
	var headers = string(splitBuffer[0])
	var body = splitBuffer[1]

	var mimetype = strings.Split(strings.Split(headers, "Content-Type: ")[1], ";")[0]
	contentLength, _ := strconv.Atoi(strings.Split(strings.Split(headers, "Content-Length: ")[1], "\r\n")[0])

	for len(body) < contentLength {
		buffer := make([]byte, 1024)
		_, err := conn.Read(buffer)
		if err != nil {
			fmt.Println("Read error: ", err.Error())
		}
		body = append(body, buffer...)
	}

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
					comment.Message = strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(string(subContent), "&", "&amp;"), "<", "&lt;"), ">", "&gt;")
				} else if strings.Contains(subHeader, `name="upload"`) {
					comment.Image = strings.ReplaceAll(strings.Split(strings.Split(subHeader, "filename=")[1], "\r\n")[0], `"`, "")
					saveFile(comment.Image, subContent)
					fmt.Println(len(loadFile(comment.Image)))
				}
			}
		}
		addComment(comment)
	}
}
