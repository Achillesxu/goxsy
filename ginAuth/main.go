// Package ginAuth
// Time    : 2023/2/2 16:15
// Author  : xushiyin
// contact : yuqingxushiyin@gmail.com

// Recipes API
//
// This is a sample recipes API. You can find out more about the API at https://github.com/PacktPublishing/Building-Distributed-Applications-in-Gin.
//
//		Schemes: http
//	 Host: localhost:8080
//		BasePath: /
//		Version: 1.0.0
//		Contact: Achillesxu <yuqingxushiyin@gmail.com>
//
//		Consumes:
//		- application/json
//
//		Produces:
//		- application/json
//
// swagger:meta
package main

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"goxsy/ginAuth/handlers"
	"goxsy/ginAuth/middlewares"
	"log"
	"os"
)

var authHandler *handlers.AuthHandler
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
	collectionUsers := client.Database(
		os.Getenv("MONGO_DATABASE")).Collection("users")
	redisClient := redis.NewClient(&redis.Options{
		Addr:     "192.168.202.158:6379",
		Password: "",
		DB:       0,
	})
	status := redisClient.Ping(context.TODO())
	fmt.Println(status)
	recipesHandler = handlers.NewRecipesHandler(ctx, collection, redisClient)
	authHandler = handlers.NewAuthHandler(ctx, collectionUsers)
}

func main() {
	r := gin.Default()
	// signin to get jwt
	r.POST("/signin", authHandler.SignInHandler)
	r.POST("/refresh", authHandler.RefreshHandler)

	auth := r.Group("/")
	auth.Use(middlewares.AuthMiddleware())
	{
		auth.POST("/recipes", recipesHandler.NewRecipesHandler)
		auth.PUT("/recipes/:id", recipesHandler.UpdateRecipesHandler)
		auth.DELETE("/recipes/:id", recipesHandler.DeleteRecipeHandler)
		auth.GET("/recipes/:id", recipesHandler.GetOneRecipeHandler)
	}
	r.GET("/recipes", recipesHandler.ListRecipesHandler)

	err := r.Run(":8080")
	if err != nil {
		log.Fatal(err)
	}
}
