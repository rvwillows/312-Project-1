package main

import (
	"bytes"
	"fmt"
	"math/rand"
	"net"
	"os"
	"strconv"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

var tokens []string = nil

func StringSliceContains(slice []string, str string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}

func loadFile(path string, visits int) []byte {
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
				if comment.Image != "" {
					commentString = commentString + "<br>" + comment.Image + "</br>"
					commentString = commentString + "<br> <img src=" + comment.Id.Hex() + ".jpg" + "> </br>"
				}
				commentString = commentString + "<br>" + comment.Message + "</br>"
			}
			dat = []byte(strings.Replace(string(dat), "{{data}}", commentString, 1))
			var token = fmt.Sprint((rand.Int63()))
			tokens = append(tokens, token)
			dat = []byte(strings.Replace(string(dat), "GOOSE12345", token, 1))
			dat = []byte(strings.Replace(string(dat), "{{pageVisits}}", fmt.Sprint(visits), 1))
		}
		return dat
	} else {
		fmt.Println("File read error 1: ", err.Error())
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

func makeResponse(status string, mimetype string, cookies []string, content []byte) []byte {
	var length = 0
	if mimetype != "" {
		length = len(content)
	}
	var response []byte = []byte(status)
	response = append(response, []byte(cr)...)
	if mimetype != "" {
		response = append(response, []byte(mimetype)...)
		response = append(response, []byte(cr)...)
		response = append(response, []byte(noSniff)...)
		response = append(response, []byte(cr)...)
	}
	for _, c := range cookies {
		response = append(response, []byte("Set-Cookie: "+c)...)
		response = append(response, []byte(cr)...)
	}
	if length != 0 {
		response = append(response, []byte("Content-Length: "+strconv.FormatInt(int64(length), 10))...)
		response = append(response, []byte(cr)...)
	}
	if mimetype != "" {
		response = append(response, []byte(cr)...)
	}
	response = append(response, content...)
	if mimetype == "" {
		response = append(response, []byte(cr)...)
		response = append(response, []byte(cr)...)
	}
	return response
}

func parseRequest(buffer []byte, conn net.Conn, req []string) {
	splitBuffer := bytes.SplitN(buffer, []byte("\r\n\r\n"), 2)
	var headers = string(splitBuffer[0])
	var body = []byte{}
	if len(splitBuffer) < 2 {
		buffer2 := make([]byte, 1024)
		_, err := conn.Read(buffer2)
		if err != nil {
			fmt.Println("Read error 2: ", err.Error())
		}
		buffer = append(buffer, buffer2...)
		splitBuffer := bytes.SplitN(buffer, []byte("\r\n\r\n"), 2)
		headers = string(splitBuffer[0])
		body = splitBuffer[1]
	} else {
		body = splitBuffer[1]
	}
	var mimetype = strings.Split(strings.Split(headers, "Content-Type: ")[1], ";")[0]
	contentLength, _ := strconv.Atoi(strings.Split(strings.Split(headers, "Content-Length: ")[1], "\r\n")[0])

	for len(body) < contentLength {
		buffer := make([]byte, 1024)
		_, err := conn.Read(buffer)
		if err != nil {
			fmt.Println("Read error 2: ", err.Error())
		}
		body = append(body, buffer...)
	}

	if mimetype == "multipart/form-data" {
		if strings.Contains(req[0], "image-upload") {
			comment := Comment{}
			var boundary = strings.Split(strings.Split(headers, "boundary=")[1], "\r\n")[0]
			content := bytes.Split(body, []byte("--"+boundary))
			var image []byte = nil
			for _, c := range content {
				if bytes.Contains(c, []byte("\r\n\r\n")) {
					subBytes := bytes.SplitN(c, []byte("\r\n\r\n"), 2)
					var subHeader = string(subBytes[0])
					var subContent = subBytes[1]
					if strings.Contains(subHeader, `name="comment"`) {
						comment.Message = strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(string(subContent), "&", "&amp;"), "<", "&lt;"), ">", "&gt;")
					} else if strings.Contains(subHeader, `name="xsrf_token"`) {

						if !StringSliceContains(tokens, strings.ReplaceAll(string(subContent), "\r\n", "")) {
							var response = makeResponse(forbidden, types["txt"], nil, loadFile("403.txt", 0))
							conn.Write([]byte(response))
							return
						}
					} else if strings.Contains(subHeader, `name="upload"`) {
						comment.Image = strings.ReplaceAll(strings.Split(strings.Split(subHeader, "filename=")[1], "\r\n")[0], `"`, "")
						image = subContent
					}
				}
			}
			id := addComment(comment)
			saveFile(id+".jpg", image)
		} else if strings.Contains(req[0], "register-form") {
			user := User{}
			var boundary = strings.Split(strings.Split(headers, "boundary=")[1], "\r\n")[0]
			content := bytes.Split(body, []byte("--"+boundary))
			for _, c := range content {
				if bytes.Contains(c, []byte("\r\n\r\n")) {
					subBytes := bytes.SplitN(c, []byte("\r\n\r\n"), 2)
					var subHeader = string(subBytes[0])
					var subContent = subBytes[1]
					if strings.Contains(subHeader, `name="username"`) {
						user.Username = strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(string(subContent), "&", "&amp;"), "<", "&lt;"), ">", "&gt;")
					} else if strings.Contains(subHeader, `name="password"`) {
						user.Password = string(subContent)
					}
				}
			}
			hash, _ := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.MinCost)
			user.Password = string(hash)
			addUser(user)
		} else if strings.Contains(req[0], "login-form") {
			user := User{}
			var boundary = strings.Split(strings.Split(headers, "boundary=")[1], "\r\n")[0]
			content := bytes.Split(body, []byte("--"+boundary))
			for _, c := range content {
				if bytes.Contains(c, []byte("\r\n\r\n")) {
					subBytes := bytes.SplitN(c, []byte("\r\n\r\n"), 2)
					var subHeader = string(subBytes[0])
					var subContent = subBytes[1]
					if strings.Contains(subHeader, `name="username"`) {
						user.Username = strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(string(subContent), "&", "&amp;"), "<", "&lt;"), ">", "&gt;")
					} else if strings.Contains(subHeader, `name="password"`) {
						user.Password = string(subContent)
					}
				}
			}
			var user2 = getUserByName(user.Username)
			err := bcrypt.CompareHashAndPassword([]byte(user2.Password), []byte(user.Password))
			if err != nil {
				var response = makeResponse(forbidden, types["txt"], nil, loadFile("403.txt", 0))
				conn.Write([]byte(response))
				return
			}
			var token = []byte(fmt.Sprint((rand.Int63())))
			hash, _ := bcrypt.GenerateFromPassword(token, bcrypt.MinCost)
			addToken(Token{Username: user2.Username, Token: string(hash)})
			var content2 = []byte("Location:  /")
			var response = makeResponse(moved, "", []string{"Token=" + string(token) + "; Max-Age=3600; HttpOnly"}, content2)
			conn.Write([]byte(response))
		}
	}
}
