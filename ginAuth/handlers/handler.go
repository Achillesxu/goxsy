// Package handlers
// Time    : 2023/2/2 10:48
// Author  : xushiyin
// contact : yuqingxushiyin@gmail.com
package handlers

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/golang-jwt/jwt/v4"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"goxsy/ginAuth/models"
	"log"
	"net/http"
	"os"
	"time"
)

type AuthHandler struct {
	collection *mongo.Collection
	ctx        context.Context
}

type Claims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

type JWTOutput struct {
	Token   string    `json:"token"`
	Expires time.Time `json:"expires"`
}

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

// swagger:operation GET /recipes recipes listRecipes
// Returns list of recipes
// ---
// produces:
// - application/json
// responses:
//
//	'200':
//	    description: Successful operation
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

// swagger:operation POST /recipes recipes newRecipe
// Create a new recipe
// ---
// produces:
// - application/json
// responses:
//
//	'200':
//	    description: Successful operation
//	'400':
//	    description: Invalid input
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

// swagger:operation PUT /recipes/{id} recipes updateRecipe
// Update an existing recipe
// ---
// parameters:
//   - name: id
//     in: path
//     description: ID of the recipe
//     required: true
//     type: string
//
// produces:
// - application/json
// responses:
//
//	'200':
//	    description: Successful operation
//	'400':
//	    description: Invalid input
//	'404':
//	    description: Invalid recipe ID
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
		{"$set", bson.D{
			{"name", recipe.Name},
			{"instructions", recipe.Instructions},
			{"ingredients", recipe.Ingredients},
			{"tags", recipe.Tags},
		}},
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

// swagger:operation DELETE /recipes/{id} recipes deleteRecipe
// Delete an existing recipe
// ---
// produces:
// - application/json
// parameters:
//   - name: id
//     in: path
//     description: ID of the recipe
//     required: true
//     type: string
//
// responses:
//
//	'200':
//	    description: Successful operation
//	'404':
//	    description: Invalid recipe ID
func (h *RecipesHandler) DeleteRecipeHandler(c *gin.Context) {
	id := c.Param("id")
	objectId, _ := primitive.ObjectIDFromHex(id)
	_, err := h.collection.DeleteOne(h.ctx, bson.M{
		"_id": objectId,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Recipe has been deleted"})

}

// swagger:operation GET /recipes/{id} recipes
// Get one recipe
// ---
// produces:
// - application/json
// parameters:
//   - name: id
//     in: path
//     description: recipe ID
//     required: true
//     type: string
//
// responses:
//
//	'200':
//	    description: Successful operation
func (h *RecipesHandler) GetOneRecipeHandler(c *gin.Context) {
	id := c.Param("id")
	objectId, _ := primitive.ObjectIDFromHex(id)
	cur := h.collection.FindOne(h.ctx, bson.M{
		"_id": objectId,
	})
	var recipe models.Recipe
	err := cur.Decode(&recipe)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, recipe)
}

func NewAuthHandler(ctx context.Context, collection *mongo.Collection) *AuthHandler {
	return &AuthHandler{
		collection: collection,
		ctx:        ctx,
	}
}

func (a *AuthHandler) SignInHandler(c *gin.Context) {
	var user models.User
	err := c.ShouldBindJSON(&user)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	h := sha256.New()
	cur := a.collection.FindOne(a.ctx, bson.M{
		"username": user.Username,
		"password": string(h.Sum([]byte(user.Password))),
	})
	if cur.Err() != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "invalid username or password",
		})
		return
	}

	expiresTime := time.Now().Add(10 * time.Minute)
	claims := &Claims{
		Username: user.Username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: &jwt.NumericDate{
				Time: expiresTime,
			},
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		c.JSON(http.StatusInternalServerError,
			gin.H{"error": err.Error()})
		return
	}
	jwtOutput := JWTOutput{
		Token:   tokenStr,
		Expires: expiresTime,
	}
	c.JSON(http.StatusOK, jwtOutput)
	return
}

func (a *AuthHandler) RefreshHandler(c *gin.Context) {
	tokenVal := c.GetHeader("Authorization")
	claims := &Claims{}
	tkn, err := jwt.ParseWithClaims(tokenVal, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWT_SECRET")), nil
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError,
			gin.H{"error": err.Error()})
		return
	}

	if tkn == nil || !tkn.Valid {
		c.JSON(http.StatusInternalServerError,
			gin.H{"error": "invalid token"})
		return
	}

	if claims.ExpiresAt.Time.Sub(time.Now()) > 30*time.Second {
		c.JSON(http.StatusBadRequest, gin.H{"error": "token is not expired yet"})
		return

	}
	expirationTime := time.Now().Add(5 * time.Minute)
	claims.ExpiresAt.Time = expirationTime
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString(os.Getenv("JWT_SECRET"))
	if err != nil {
		c.JSON(http.StatusInternalServerError,
			gin.H{"error": err.Error()})
		return
	}
	jwtOut := JWTOutput{
		Token:   tokenStr,
		Expires: expirationTime,
	}
	c.JSON(http.StatusOK, jwtOut)
}
