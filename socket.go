package main

import (
	"crypto/sha1"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math/rand"
	"net"
	"strings"
)

var actitveConnections map[string]net.Conn

func webSocketHandshake(conn net.Conn, req []string) []byte {
	var index = 0
	for i, s := range req {
		if strings.HasPrefix(s, "Sec-WebSocket-Key: ") {
			index = i
		}
	}
	var key = strings.Split(req[index], "Sec-WebSocket-Key: ")[1]
	var hash = sha1.Sum([]byte(key + GUID))
	var accept = base64.StdEncoding.EncodeToString(hash[:])

	var content = "Connection: Upgrade" + cr + "Upgrade: websocket" + cr + "Sec-WebSocket-Accept: " + accept

	var response = makeResponse(switching, "", []byte(content))
	return response
}

func webSocketServer(conn net.Conn) {
	var username = "Goose#" + fmt.Sprint(rand.Intn(9999))
	if actitveConnections == nil {
		actitveConnections = make(map[string]net.Conn)
	}
	actitveConnections[username] = conn
	for {
		buffer := make([]byte, 1024)
		_, err := conn.Read(buffer)
		if err != nil {
			fmt.Println("Read error: ", err.Error())
		}
		var opcode = buffer[0] & 15
		var maskBit = (buffer[1] & 128) / 128
		var length = int(buffer[1] & 127)
		var mask []byte
		var payloadIndex = 2
		if length == 126 {
			//Read the next 16 bits
			length = int(binary.BigEndian.Uint16(buffer[2:4]))
			payloadIndex = 4

		} else if length == 127 {
			//Read the next 64 bits
			length = int(binary.BigEndian.Uint64(buffer[2:10]))
			payloadIndex = 10
		}

		if length > 1024 {
			for length > len(buffer) {
				newBuffer := make([]byte, 1024)
				_, err := conn.Read(newBuffer)
				if err != nil {
					fmt.Println("Read error: ", err.Error())
				}
				buffer = append(buffer, newBuffer...)
			}
		}

		if maskBit == 1 {
			mask = buffer[payloadIndex : payloadIndex+4]
			payloadIndex = payloadIndex + 4

		}

		var payload = buffer[payloadIndex : payloadIndex+length]

		if maskBit == 1 {
			var counter = 0
			for i, _ := range payload {
				payload[i] = payload[i] ^ mask[counter]
				counter = counter + 1
				if counter == 4 {
					counter = 0
				}
			}
		}

		if opcode == 8 {
			delete(actitveConnections, username)
			var res = makeFrame([]byte(""))
			conn.Write(res)

			return
		}
		if opcode == 1 {
			// Format is text
			var result map[string]string
			json.Unmarshal(payload, &result)

			if result["messageType"] == "chatMessage" {
				var messageText = strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(result["comment"], "&", "&amp;"), "<", "&lt;"), ">", "&gt;")
				messageObject := new(Message)
				messageObject.Message = messageText
				messageObject.Username = username
				addMessage(*messageObject)

				message := map[string]string{"messageType": "chatMessage", "username": username, "comment": messageText}
				response, _ := json.Marshal(&message)
				var res = makeFrame(response)
				for _, c := range actitveConnections {
					c.Write(res)
				}
			} else if result["messageType"] == "webRTC-offer" || result["messageType"] == "webRTC-answer" || result["messageType"] == "webRTC-candidate" {
				var frame = makeFrame(payload)
				for u, c := range actitveConnections {
					if u != username {
						c.Write(frame)
					}
				}
			}
		}
		if opcode == 2 {
			//Format is binary
		}
	}
}

func makeFrame(payload []byte) []byte {
	var length = len(payload)
	var payloadIndex = 2

	var sizeLength = 0
	if length < 126 {
		sizeLength = 2
	} else if length >= 126 && length < 65536 {
		sizeLength = 4

	} else if length >= 65536 {
		sizeLength = 10
	}

	frame := make([]byte, length+sizeLength)
	if length == 0 {
		frame[0] = 136
	} else {
		frame[0] = 129
	}

	if length < 126 {
		frame[1] = byte(length)
	} else if length >= 126 && length < 65536 {
		frame[1] = 126
		binary.BigEndian.PutUint16(frame[2:4], uint16(length))
		payloadIndex = 4

	} else if length >= 65536 {
		frame[1] = 127
		binary.BigEndian.PutUint64(frame[2:10], uint64(length))
		payloadIndex = 10
	}

	for _, b := range payload {
		frame[payloadIndex] = b
		payloadIndex = payloadIndex + 1
	}

	return frame
}
