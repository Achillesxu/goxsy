// Package handlers
// Time    : 2023/2/2 10:48
// Author  : xushiyin
// contact : yuqingxushiyin@gmail.com
package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"goxsy/ginWithMongo/models"
	"log"
	"net/http"
	"time"
)

type RecipesHandler struct {
	collection  *mongo.Collection
	ctx         context.Context
	redisClient *redis.Client
}

func NewRecipesHandler(
	ctx context.Context,
	collection *mongo.Collection,
	redisClient *redis.Client) *RecipesHandler {
	return &RecipesHandler{
		collection:  collection,
		ctx:         ctx,
		redisClient: redisClient,
	}
}

func (h *RecipesHandler) ListRecipesHandler(c *gin.Context) {

	recipes := make([]models.Recipe, 0)
	val, err := h.redisClient.Get(h.ctx, "recipes").Result()
	if err == redis.Nil {
		log.Printf("request to mongodb")
		cur, err := h.collection.Find(h.ctx, bson.M{})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer cur.Close(h.ctx)

		for cur.Next(h.ctx) {
			var recipe models.Recipe
			err := cur.Decode(&recipe)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			recipes = append(recipes, recipe)
			data, _ := json.Marshal(recipes)
			h.redisClient.Set(h.ctx, "recipes", string(data), 0)
		}
	} else if err != nil {
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	} else {
		log.Printf("request to redis")
		json.Unmarshal([]byte(val), &recipes)
	}
	c.JSON(http.StatusOK, recipes)
}

func (h *RecipesHandler) NewRecipesHandler(c *gin.Context) {
	var recipe models.Recipe
	err := c.ShouldBindJSON(&recipe)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	recipe.ID = primitive.NewObjectID()
	recipe.PublishedAt = time.Now()
	_, err = h.collection.InsertOne(h.ctx, recipe)
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Err while inserting a new recipe: " + err.Error()})
		return
	}
	log.Println("remove old recipes from redis")
	h.redisClient.Del(h.ctx, "recipes")
	c.JSON(http.StatusOK, recipe)
}

func (h *RecipesHandler) UpdateRecipesHandler(c *gin.Context) {
	id := c.Param("id")
	var recipe models.Recipe
	err := c.ShouldBindJSON(&recipe)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	objId, _ := primitive.ObjectIDFromHex(id)
	_, err = h.collection.UpdateOne(h.ctx, bson.M{"_id": objId}, bson.D{
		{"name", recipe.Name},
		{"instructions", recipe.Instructions},
		{"ingredients", recipe.Ingredients},
		{"tags", recipe.Tags},
	})
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	log.Println("remove old recipes from redis")
	h.redisClient.Del(h.ctx, "recipes")
	c.JSON(http.StatusOK, gin.H{"message": "recipe has been updated"})
}
