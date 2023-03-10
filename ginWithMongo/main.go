// Package ginWithMongo
// Time    : 2023/1/31 11:10
// Author  : xushiyin
// contact : yuqingxushiyin@gmail.com
package main

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"goxsy/ginWithMongo/handlers"
	"log"
	"os"
)

var recipesHandler *handlers.RecipesHandler

func init() {
	ctx := context.Background()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(os.Getenv("MONGO_URI")))
	if err = client.Ping(context.TODO(), readpref.Primary()); err != nil {
		log.Fatal(err)
	}
	log.Println("Connected to MongoDB")
	collection := client.Database(
		os.Getenv("MONGO_DATABASE")).Collection("recipes")
	redisClient := redis.NewClient(&redis.Options{
		Addr: "192.168.202.158:6379",

		Password: "",
		DB:       0,
	})
	status := redisClient.Ping(context.TODO())
	fmt.Println(status)
	recipesHandler = handlers.NewRecipesHandler(ctx, collection, redisClient)
}

func main() {
	r := gin.Default()
	r.POST("/recipes", recipesHandler.NewRecipesHandler)
	r.GET("/recipes", recipesHandler.ListRecipesHandler)
	r.PUT("/recipes/:id", recipesHandler.UpdateRecipesHandler)
	err := r.Run()
	if err != nil {
		log.Fatal(err)
	}
}
