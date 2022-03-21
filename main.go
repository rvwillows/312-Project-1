package main

import (
	"bytes"
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
	} else if strings.HasPrefix(req[0], "POST") {
		postHandler(conn, req)
	} else if strings.HasPrefix(req[0], "PUT") {
		putHandler(conn, req)
	} else if strings.HasPrefix(req[0], "DELETE") {
		deleteHandler(conn, req)
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
		if content == nil {
			status = notFound
			mimetype = types["txt"]
			content = loadFile("404.txt")
		}
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

			if path == "$users" {
				content, err := json.Marshal(getUsers())
				if err != nil {
					log.Fatal(err)
				}
				var response []byte = makeResponse(status, mimetype, content)
				return response
			}

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
		if content == nil {
			status = notFound
			mimetype = types["txt"]
			content = loadFile("404.txt")
		}
		var response []byte = makeResponse(status, mimetype, content)
		return response
	}
}

func getHandler(conn net.Conn, req []string) {
	var path string = strings.Split(req[0], " ")[1]
	for _, point := range exposed {
		if strings.HasPrefix(path, point) {
			var id = strings.TrimLeft(path, point)
			if id != "" {
				var user = getUser(id)
				if user.Id == nil {
					var response []byte = contentResolve("404")
					conn.Write([]byte(response))
					return
				}
				content, err := json.Marshal(user)
				if err != nil {
					log.Fatal(err)
				}
				var response = makeResponse(ok, types["json"], content)
				conn.Write([]byte(response))
				return
			}
		}
	}
	var response []byte = contentResolve(path)
	conn.Write([]byte(response))
}

func postHandler(conn net.Conn, req []string) {
	var long_path string = strings.Split(req[0], " ")[1]
	var path = strings.Split(long_path, "?")[0]
	var values string = req[len(req)-1]
	var action = router[path]
	var response []byte
	for _, point := range exposed {
		if strings.HasPrefix(path, point) {
			var id = strings.TrimLeft(path, point)
			if id != "" {
				return
			}
		}
	}
	if action == "$users" {
		user := User{}
		err := json.Unmarshal(bytes.Trim([]byte(values), "\x00"), &user)
		if err != nil {
			log.Fatal(err)
		}
		content, err := json.Marshal(addUser(user))
		if err != nil {
			log.Fatal(err)
		}
		response = makeResponse(created, types["json"], content)
	}
	conn.Write([]byte(response))
}

func putHandler(conn net.Conn, req []string) {
	var path string = strings.Split(req[0], " ")[1]
	var values string = req[len(req)-1]
	for _, point := range exposed {
		if strings.HasPrefix(path, point) {
			var id = strings.TrimLeft(path, point)
			if id != "" {
				user := User{}
				err := json.Unmarshal(bytes.Trim([]byte(values), "\x00"), &user)
				if err != nil {
					log.Fatal(err)
				}
				var updatedUser = updateUser(user, id)
				if updatedUser.Id == nil {
					var response []byte = contentResolve("404")
					conn.Write([]byte(response))
					return
				}
				content, err := json.Marshal(updatedUser)
				if err != nil {
					log.Fatal(err)
				}
				var response = makeResponse(ok, types["json"], content)
				conn.Write([]byte(response))
				return
			}
		}
	}
}

func deleteHandler(conn net.Conn, req []string) {
	var path string = strings.Split(req[0], " ")[1]
	for _, point := range exposed {
		if strings.HasPrefix(path, point) {
			var id = strings.TrimLeft(path, point)
			if id != "" {
				if deleteUser(id) {
					var response = makeResponse(noContent, types["json"], []byte(nil))
					conn.Write([]byte(response))
					return
				} else {
					var response []byte = contentResolve("404")
					conn.Write([]byte(response))
					return
				}
			}
		}
	}
}
