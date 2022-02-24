package main

import (
	"math/big"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

var router = map[string]string{
	"/hello": "hello.txt",
	"/hi":    "/hello",
	"/":      "index.html",
	"/users": "$users",
}

var exposed = []string{
	"/users",
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
	Email    string             `json:"email"`
	Username string             `json:"username"`
}

type UserButBetter struct {
	Id       *big.Int `bson:"_id" json:"id,omitempty"`
	Email    string   `json:"email"`
	Username string   `json:"username"`
}

var ok = "HTTP/1.1 200 OK"

var created = "HTTP/1.1 201 Created"

var noContent = "HTTP/1.1 204 No Content"

var moved = "HTTP/1.1 302 Temporarily Moved"

var notFound = "HTTP/1.1 404 Not Found"

var noSniff = "X-Content-Type-Options: nosniff"

var cr = "\r\n"
