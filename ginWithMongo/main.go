// Package ginWithMongo
// Time    : 2023/1/31 11:10
// Author  : xushiyin
// contact : yuqingxushiyin@gmail.com
package main

import (
	"context"
	"github.com/gin-gonic/gin"
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
	recipesHandler = handlers.NewRecipesHandler(ctx, collection)
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
