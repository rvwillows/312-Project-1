package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"strconv"
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
		fmt.Println("Read error: 3", err.Error())
	}

	req := strings.Split(string(buffer), "\r\n")
	if strings.HasPrefix(req[0], "GET") {
		getHandler(conn, req)
	} else if strings.HasPrefix(req[0], "POST") {
		parseRequest(buffer, conn, req)
		postHandler(conn, req)
	} else if strings.HasPrefix(req[0], "PUT") {
		putHandler(conn, req)
	} else if strings.HasPrefix(req[0], "DELETE") {
		deleteHandler(conn, req)
	}
	conn.Close()
}

func contentResolve(path string, cookies map[string]string) []byte {
	var status string = ""
	var mimetype string = ""
	var content []byte = nil

	// Check if it's a file
	if strings.Contains(path, ".") {
		status = ok
		var split = strings.Split(path, ".")
		mimetype = types[split[len(split)-1]]
		content = loadFile(path, 0)
		if content == nil {
			status = notFound
			mimetype = types["txt"]
			content = loadFile("404.txt", 0)
		}
		var response []byte = makeResponse(status, mimetype, nil, content)
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
				var response []byte = makeResponse(status, mimetype, nil, content)
				return response
			}
		} else if strings.HasPrefix(path, "/") {
			status = moved
			content = []byte("Location: " + path)
			var response []byte = makeResponse(status, mimetype, nil, content)
			return response
			// Otherwise, its a file and we can simply load in
		} else if strings.HasPrefix(path, "index.html") {
			visits, _ := strconv.Atoi(cookies["visits"])
			visits = visits + 1
			status = ok
			var split = strings.Split(path, ".")
			mimetype = types[split[len(split)-1]]
			content = loadFile(path, visits)
			var response []byte = makeResponse(status, mimetype, []string{"visits=" + fmt.Sprint(visits) + "; Max-Age=3600; Path=/"}, content)
			return response
		} else {
			status = ok
			var split = strings.Split(path, ".")
			mimetype = types[split[len(split)-1]]
		}
		content = loadFile(path, 0)
		if content == nil {
			status = notFound
			mimetype = types["txt"]
			content = loadFile("404.txt", 0)
		}
		var response []byte = makeResponse(status, mimetype, nil, content)
		return response
	}
}

func getHandler(conn net.Conn, req []string) {
	var path string = strings.Split(req[0], " ")[1]
	cookies := make(map[string]string)
	for _, s := range req {
		if strings.Contains(s, "Cookie") {
			var cookieList = strings.Split(strings.Replace(s, "Cookie: ", "", 1), "; ")
			for _, s := range cookieList {
				var cookie = strings.Split(s, "=")
				cookies[cookie[0]] = cookie[1]
			}
		}
	}
	if path == "/websocket" {
		var response = webSocketHandshake(conn, req)
		conn.Write(response)
		webSocketServer(conn)
	} else if path == "/chat-history" {
		content, err := json.Marshal(getMessages())
		if err != nil {
			log.Fatal(err)
		}
		var response = makeResponse(ok, types["json"], nil, content)
		conn.Write([]byte(response))
		return
	} else {
		for _, point := range exposed {
			if strings.HasPrefix(path, point) {
				var id = strings.TrimLeft(path, point)
				if id != "" {
					var user = getUser(id)
					if user.Id == nil {
						var response []byte = contentResolve("404", nil)
						conn.Write([]byte(response))
						return
					}
					content, err := json.Marshal(user)
					if err != nil {
						log.Fatal(err)
					}
					var response = makeResponse(ok, types["json"], nil, content)
					conn.Write([]byte(response))
					return
				}
			}
		}
		var response []byte = contentResolve(path, cookies)
		conn.Write([]byte(response))
	}
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
		response = makeResponse(created, types["json"], nil, content)
	}
	if action == "$image-upload" {
		var content = []byte("Location:  /")
		response = makeResponse(moved, "", nil, content)
	}
	if action == "$register" {
		var content = []byte("Location:  /")
		response = makeResponse(moved, "", nil, content)
	}
	if action == "$login" {
		var content = []byte("Location:  /")
		response = makeResponse(moved, "", nil, content)
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
					var response []byte = contentResolve("404", nil)
					conn.Write([]byte(response))
					return
				}
				content, err := json.Marshal(updatedUser)
				if err != nil {
					log.Fatal(err)
				}
				var response = makeResponse(ok, types["json"], nil, content)
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
					var response = makeResponse(noContent, types["json"], nil, []byte(nil))
					conn.Write([]byte(response))
					return
				} else {
					var response []byte = contentResolve("404", nil)
					conn.Write([]byte(response))
					return
				}
			}
		}
	}
}
