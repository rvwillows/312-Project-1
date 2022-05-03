package main

import (
	"context"
	"fmt"
	"log"
	"math/big"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var UserCollection *mongo.Collection
var CommentCollection *mongo.Collection
var MessageCollection *mongo.Collection
var TokenCollection *mongo.Collection
var ctx = context.Background()

func connect(uri string) {
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatal(err)
	}
	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}
	database := client.Database("website")
	UserCollection = database.Collection("users")
	CommentCollection = database.Collection("comments")
	MessageCollection = database.Collection("message")
	TokenCollection = database.Collection("token")
	fmt.Println("Connected to MongoDB!")
}

func getUser(id string) User {
	userId := new(big.Int)
	userId.SetString(id, 10)

	id = fmt.Sprintf("%x", userId)

	var result User
	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return User{}
	}

	err = UserCollection.
		FindOne(ctx, bson.D{{Key: "_id", Value: objectId}}).
		Decode(&result)
	if err != nil {
		return User{}
	}
	user := User{Id: objectId, Password: result.Password, Username: result.Username, Salt: result.Salt}
	return user
}

func getUserByName(name string) User {
	var result User

	err := UserCollection.
		FindOne(ctx, bson.D{{Key: "username", Value: name}}).
		Decode(&result)
	if err != nil {
		return User{}
	}
	user := User{Id: result.Id, Password: result.Password, Username: result.Username, Salt: result.Salt}
	return user
}

func getTokenByName(name string) Token {
	var result Token
	err := TokenCollection.FindOne(ctx, bson.D{{Key: "username", Value: name}}).Decode(&result)
	if err != nil {
		return Token{}
	}
	return result
}

func getComments() []Comment {
	cursor, err := CommentCollection.Find(ctx, bson.M{})
	if err != nil {
		log.Fatal(err)
	}
	result := Comment{}
	comments := []Comment{}
	for cursor.Next(ctx) {
		if err = cursor.Decode(&result); err != nil {
			log.Fatal(err)
		}
		comments = append(comments, result)
	}
	return comments
}

func getMessages() []Message {
	cursor, err := MessageCollection.Find(ctx, bson.M{})
	if err != nil {
		log.Fatal(err)
	}
	result := Message{}
	messages := []Message{}
	for cursor.Next(ctx) {
		if err = cursor.Decode(&result); err != nil {
			log.Fatal(err)
		}
		messages = append(messages, result)
	}
	return messages
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
		betterUser := User{Id: result.Id, Password: result.Password, Username: result.Username, Salt: result.Salt}
		users = append(users, betterUser)
	}
	return users
}

func addComment(comment Comment) string {
	comment.Id = primitive.NewObjectID()
	result, err := CommentCollection.InsertOne(ctx, comment)
	if err != nil {
		log.Fatal(err)
	}
	objectID := result.InsertedID.(primitive.ObjectID)
	return objectID.Hex()
}

func addMessage(message Message) string {
	result, err := MessageCollection.InsertOne(ctx, message)
	if err != nil {
		log.Fatal(err)
	}
	objectID := result.InsertedID.(primitive.ObjectID)
	return objectID.Hex()
}

func addUser(user User) User {
	user.Id = primitive.NewObjectID()
	result, err := UserCollection.InsertOne(ctx, user)
	if err != nil {
		log.Fatal(err)
	}
	objectID := result.InsertedID.(primitive.ObjectID)
	betterUser := User{Id: objectID, Password: user.Password, Username: user.Username, Salt: user.Salt}
	return betterUser
}

func addToken(token Token) string {
	result, err := TokenCollection.InsertOne(ctx, token)
	if err != nil {
		log.Fatal(err)
	}
	objectID := result.InsertedID.(primitive.ObjectID)
	return objectID.Hex()
}

func updateUser(user User, id string) User {
	user.Id = primitive.NewObjectID()

	userId := new(big.Int)
	userId.SetString(id, 10)

	id = fmt.Sprintf("%x", userId)
	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return User{}
	}

	filter := bson.D{{Key: "_id", Value: objectId}}
	after := options.After

	update := bson.D{{Key: "$set", Value: bson.D{{Key: "password", Value: user.Password}, {Key: "username", Value: user.Username}, {Key: "salt", Value: user.Salt}}}}

	returnOpt := options.FindOneAndUpdateOptions{
		ReturnDocument: &after,
	}

	updateResult := UserCollection.FindOneAndUpdate(ctx, filter, update, &returnOpt)
	var result User
	_ = updateResult.Decode(&result)
	betterUser := User{Id: result.Id, Password: result.Password, Username: result.Username, Salt: result.Salt}
	return betterUser
}

func deleteUser(id string) bool {
	userId := new(big.Int)
	userId.SetString(id, 10)

	id = fmt.Sprintf("%x", userId)
	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return false
	}

	filter := bson.D{{Key: "_id", Value: objectId}}

	result, err := UserCollection.DeleteOne(ctx, filter)
	if err != nil {
		log.Fatal(err)
	}
	if result.DeletedCount != 0 {
		return true
	}
	return false
}
