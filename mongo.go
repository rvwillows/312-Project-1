package main

import (
	"context"
	"fmt"
	"log"
	"math/big"
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

func getUser(id string) UserButBetter {
	userId := new(big.Int)
	userId.SetString(id, 10)

	id = fmt.Sprintf("%x", userId)

	var result User
	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return UserButBetter{}
	}

	err = UserCollection.
		FindOne(ctx, bson.D{{Key: "_id", Value: objectId}}).
		Decode(&result)
	if err != nil {
		return UserButBetter{}
	}
	resultId := new(big.Int)
	resultId.SetString(result.Id.Hex(), 16)
	user := UserButBetter{Id: resultId, Email: result.Email, Username: result.Username}
	return user

}

func getUsers() []UserButBetter {
	cursor, err := UserCollection.Find(ctx, bson.M{})
	if err != nil {
		log.Fatal(err)
	}
	result := User{}
	users := []UserButBetter{}
	for cursor.Next(ctx) {
		if err = cursor.Decode(&result); err != nil {
			log.Fatal(err)
		}
		id := new(big.Int)
		id.SetString(result.Id.Hex(), 16)
		betterUser := UserButBetter{Id: id, Email: result.Email, Username: result.Username}
		users = append(users, betterUser)
		fmt.Println(result)
	}
	fmt.Println(users)
	return users
}

func addUser(values []string) UserButBetter {
	user := User{}
	user.Id = primitive.NewObjectID()
	user.Email = strings.TrimPrefix(values[0], "email=")
	user.Username = strings.TrimPrefix(values[1], "username=")
	fmt.Println(user.Email)
	fmt.Println(user.Username)
	result, err := UserCollection.InsertOne(ctx, user)
	if err != nil {
		log.Fatal(err)
	}
	objectID := result.InsertedID.(primitive.ObjectID)
	id := new(big.Int)
	id.SetString(objectID.Hex(), 16)
	betterUser := UserButBetter{Id: id, Email: user.Email, Username: user.Username}
	fmt.Println(betterUser)
	return betterUser
}

func updateUser(values []string, id string) UserButBetter {
	user := User{}
	user.Id = primitive.NewObjectID()
	user.Email = strings.TrimPrefix(values[0], "email=")
	user.Username = strings.TrimPrefix(values[1], "username=")
	fmt.Println(user.Email)
	fmt.Println(user.Username)

	userId := new(big.Int)
	userId.SetString(id, 10)

	id = fmt.Sprintf("%x", userId)
	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return UserButBetter{}
	}

	filter := bson.D{{Key: "_id", Value: objectId}}
	after := options.After

	update := bson.D{{Key: "$set", Value: bson.D{{Key: "email", Value: user.Email}, {Key: "username", Value: user.Username}}}}

	returnOpt := options.FindOneAndUpdateOptions{
		ReturnDocument: &after,
	}
	fmt.Println(update)
	fmt.Println(filter)

	updateResult := UserCollection.FindOneAndUpdate(ctx, filter, update, &returnOpt)
	var result User
	_ = updateResult.Decode(&result)

	newid := new(big.Int)
	newid.SetString(string(result.Id.Hex()), 16)
	betterUser := UserButBetter{Id: newid, Email: result.Email, Username: result.Username}
	fmt.Println(betterUser)
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
