// Package main
// Time    : 2023/2/6 21:24
// Author  : xushiyin
// contact : yuqingxushiyin@gmail.com
package main

import (
	"context"
	"crypto/sha256"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"log"
)

func main() {
	users := map[string]string{
		"admin":      "fCRmh4Q2J7Rseqkz",
		"packt":      "RE4zfHB35VPtTkbT",
		"mlabouardy": "L3nSFRcZzNQ67bcc",
	}

	ctx := context.Background()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://admin:admin@192.168.202.158:27017/test?authSource=admin"))
	if err = client.Ping(context.TODO(), readpref.Primary()); err != nil {
		log.Fatal(err)
	}

	collection := client.Database("demo").Collection("users")
	h := sha256.New()

	for username, password := range users {
		collection.InsertOne(ctx, bson.M{
			"username": username,
			"password": string(h.Sum([]byte(password))),
		})
	}
}
