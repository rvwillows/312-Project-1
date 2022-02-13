package main

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
)

var router = map[string]string{
	"/hello": "hello.txt",
	"/hi":    "/hello",
	"/":      "index.html",
}

var types = map[string]string{
	"txt":  "Content-Type: text/plain; charset=utf-8",
	"html": "Content-Type: text/html; charset=utf-8",
	"css":  "Content-Type: text/html; charset=utf-8",
	"js":   "Content-Type: text/js; charset=utf-8",
	"png":  "Content-Type: image/png",
	"jpg":  "Content-Type: image/jpeg",
	"jpeg": "Content-Type: image/jpeg",
	"mp4":  "Content-Type: video/mp4",
}

var ok = "HTTP/1.1 200 OK"

var moved = "HTTP/1.1 302 Temporarily Moved"

var notFound = "HTTP/1.1 404 Not Found"

var cr = "\r\n"

func main() {
	fmt.Println("Starting server!")
	ln, err := net.Listen("tcp", ":8000")
	if err != nil {
		fmt.Println("Listen error: ", err.Error())
	}
	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("Connection error: ", err.Error())
		}
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	buffer := make([]byte, 1024)
	_, err := conn.Read(buffer)
	if err != nil {
		fmt.Println("Read error: ", err.Error())
	}
	req := strings.Split(string(buffer), "\r\n")
	if strings.HasPrefix(req[0], "GET") {
		getHandler(conn, req)
	}
	if strings.HasPrefix(req[0], "POST") {
		postHandler(conn, req)
	}
	conn.Close()
}

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

func contentResolve(path string) []byte {
	var status string = ""
	var mimetype string = ""
	var length string = ""
	var content []byte = nil

	// Check if it's a file
	if strings.Contains(path, ".") {
		//Remove the first / then load the file il;n
		strings.Replace(path, "/", "", 1)
		status = ok
		var split = strings.Split(path, ".")
		mimetype = types[split[len(split)-1]]
		content = loadFile(path)
		length = "Content-Length: " + strconv.FormatInt(int64(len(content)), 10)
		var response []byte = []byte(status)
		response = append(response, []byte(cr)...)
		response = append(response, []byte(mimetype)...)
		response = append(response, []byte(cr)...)
		response = append(response, []byte(length)...)
		response = append(response, []byte(cr)...)
		response = append(response, []byte(cr)...)
		response = append(response, content...)
		return response

		// If not, then it's a path
	} else {
		path = router[path]
		// If the path is not in the router, return 404
		if path == "" {
			status = notFound
			mimetype = types["txt"]
			path = "404.txt"
			// If it's another router path, send a redirect
		} else if strings.HasPrefix(path, "/") {
			status = moved
			path = "Location: " + path
			length = "Content-Length: 0"
			var response []byte = []byte(status)
			response = append(response, []byte(cr)...)
			response = append(response, []byte(length)...)
			response = append(response, []byte(cr)...)
			response = append(response, []byte(path)...)
			response = append(response, []byte(cr)...)
			response = append(response, []byte(cr)...)
			return response
			// Otherwise, its a file and we can simply load in
		} else {
			status = ok
			var split = strings.Split(path, ".")
			mimetype = types[split[len(split)-1]]
		}
		content = loadFile(path)
		length = "Content-Length: " + strconv.FormatInt(int64(len(content)), 10)
		var response []byte = []byte(status)
		response = append(response, []byte(cr)...)
		response = append(response, []byte(mimetype)...)
		response = append(response, []byte(cr)...)
		response = append(response, []byte(length)...)
		response = append(response, []byte(cr)...)
		response = append(response, []byte(cr)...)
		response = append(response, content...)
		return response
	}
}

func getHandler(conn net.Conn, req []string) {
	var path string = strings.Split(req[0], " ")[1]
	var response []byte = contentResolve(path)
	conn.Write([]byte(response))
}

func postHandler(conn net.Conn, req []string) {

}
