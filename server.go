package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"strings"
)

func main() {

	connect("mongodb://mongo:27017")

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

func contentResolve(path string) []byte {
	var status string = ""
	var mimetype string = ""
	var content []byte = nil
	// Check if it's a file
	if strings.Contains(path, ".") {
		status = ok
		var split = strings.Split(path, ".")
		mimetype = types[split[len(split)-1]]
		content = loadFile(path)
		var response []byte = makeResponse(status, mimetype, content)
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
		} else if strings.HasPrefix(path, "$") {
			// Check if it's a database call
			status = ok
			mimetype = types["json"]

			content, err := json.Marshal(getUsers())
			if err != nil {
				log.Fatal(err)
			}
			var response []byte = makeResponse(status, mimetype, content)
			return response

		} else if strings.HasPrefix(path, "/") {
			status = moved
			content = []byte("Location: " + path)
			var response []byte = makeResponse(status, mimetype, content)
			return response
			// Otherwise, its a file and we can simply load in
		} else {
			status = ok
			var split = strings.Split(path, ".")
			mimetype = types[split[len(split)-1]]
		}
		content = loadFile(path)
		var response []byte = makeResponse(status, mimetype, content)
		return response
	}
}

func getHandler(conn net.Conn, req []string) {
	var path string = strings.Split(req[0], " ")[1]
	var response []byte = contentResolve(path)
	conn.Write([]byte(response))
}

func postHandler(conn net.Conn, req []string) {
	var long_path string = strings.Split(req[0], " ")[1]
	var path = strings.Split(long_path, "?")[0]
	var values []string = strings.Split(strings.Split(long_path, "?")[1], "&")
	var response = router[path]

	if response == "$users" {
		addUser(values)
	}
	conn.Write([]byte(created))
}
