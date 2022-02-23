package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var router = map[string]string{
	"/hello": "hello.txt",
	"/hi":    "/hello",
	"/":      "index.html",
	"/users": "$users",
}

var types = map[string]string{
	"txt":  "Content-Type: text/plain; charset=utf-8",
	"html": "Content-Type: text/html; charset=utf-8",
	"css":  "Content-Type: text/css; charset=utf-8",
	"js":   "Content-Type: text/javascript; charset=utf-8",
	"png":  "Content-Type: image/png",
	"jpg":  "Content-Type: image/jpeg",
	"jpeg": "Content-Type: image/jpeg",
	"mp4":  "Content-Type: video/mp4",
	"json": "Content-Type: application/json",
}

type User struct {
	Id       primitive.ObjectID `bson:"_id" json:"id,omitempty"`
	Username string             `json:"username"`
	Email    string             `json:"email"`
}

var ok = "HTTP/1.1 200 OK"

var created = "HTTP/1.1 201 Created"

var moved = "HTTP/1.1 302 Temporarily Moved"

var notFound = "HTTP/1.1 404 Not Found"

var noSniff = "X-Content-Type-Options: nosniff"

var cr = "\r\n"

var UserCollection *mongo.Collection
var ctx = context.Background()

func main() {
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://mongo:27017"))
	if err != nil {
		log.Fatal(err)
	}
	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}
	database := client.Database("website")
	UserCollection = database.Collection("users")
	fmt.Println("Connected to MongoDB!")

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
		status = ok
		var split = strings.Split(path, ".")
		mimetype = types[split[len(split)-1]]
		content = loadFile(path)
		length = "Content-Length: " + strconv.FormatInt(int64(len(content)), 10)
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
			length = "Content-Length: " + strconv.FormatInt(int64(len(content)), 10)
			var response []byte = []byte(status)
			response = append(response, []byte(cr)...)
			response = append(response, []byte(mimetype)...)
			response = append(response, []byte(cr)...)
			response = append(response, []byte(length)...)
			response = append(response, []byte(cr)...)
			response = append(response, []byte(cr)...)
			response = append(response, content...)
			fmt.Println(string(response))
			return response

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
	var long_path string = strings.Split(req[0], " ")[1]
	var path = strings.Split(long_path, "?")[0]
	var values []string = strings.Split(strings.Split(long_path, "?")[1], "&")
	var response = router[path]

	if response == "$users" {
		addUser(values)
	}
	conn.Write([]byte(created))
}

func getUsers() []User {
	cursor, err := UserCollection.Find(ctx, bson.M{})
	if err != nil {
		log.Fatal(err)
	}
	result := User{}
	users := []User{}
	for cursor.Next(ctx) {
		if err = cursor.Decode(&result); err != nil {
			log.Fatal(err)
		}
		users = append(users, result)
		fmt.Println(result)
	}
	fmt.Println(users)
	return users
}

func addUser(values []string) {
	user := User{}
	user.Id = primitive.NewObjectID()
	user.Email = strings.TrimPrefix(values[0], "email=")
	user.Username = strings.TrimPrefix(values[1], "username=")
	fmt.Println(user.Email)
	fmt.Println(user.Email)
	result, err := UserCollection.InsertOne(ctx, user)
	if err != nil {
		log.Fatal(err)
	}
	objectID := result.InsertedID.(primitive.ObjectID)
	fmt.Println(objectID)
}
