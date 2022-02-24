package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var UserCollection *mongo.Collection
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
	fmt.Println("Connected to MongoDB!")
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
