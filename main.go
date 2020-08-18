package main

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	r := gin.Default()
	mongoContext, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	mongoAdmin := os.Getenv("USER")
	mongoPassword := os.Getenv("PASSWORD")

	mongoUri := fmt.Sprintf(
		"mongodb+srv://%s:%s@golang-docker-test-clus.mdqzh.gcp.mongodb.net/test_database?retryWrites=true&w=majority",
		mongoAdmin, mongoPassword,
	)

	client, clientError := mongo.Connect(mongoContext, options.Client().ApplyURI(mongoUri))
	if clientError != nil {
		panic(clientError)
	}

	r.GET("/", func(c *gin.Context) {
		c.String(200, "Hello there stranger!")
	})
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
	r.POST("/fruits", func(c *gin.Context) {
		writeContext, writeContextCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer writeContextCancel()
		fruits := map[string]string{}
		bindError := c.ShouldBindJSON(&fruits)

		if bindError != nil {
			c.JSON(http.StatusUnprocessableEntity, map[string]interface{}{
				"success": true,
				"message": "invalid json body",
			})
			return
		}

		if len(fruits) == 0 {
			c.JSON(http.StatusBadRequest, map[string]interface{}{
				"success": true,
				"message": "no fruits specified",
			})
		}

		var values []interface{}
		for key, value := range fruits {
			values = append(values, bson.M{
				key:   key,
				value: value,
			})
		}
		fruitsCollection := client.Database("test_database").Collection("fruits")
		result, insertionError := fruitsCollection.InsertMany(writeContext, values)
		if insertionError != nil {
			log.Print(insertionError.Error())
			c.JSON(http.StatusInternalServerError, map[string]interface{}{
				"success": true,
				"message": "could not insert fruits",
			})
			return
		}
		c.JSON(http.StatusCreated, map[string]interface{}{
			"success": true,
			"data":    result.InsertedIDs,
			"message": "fruits inserted",
		})

	})
	log.Fatal(r.Run(":8080"))
}
