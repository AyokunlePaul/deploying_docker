package main

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"log"
	"net/http"
	"os"
	"time"
)

const (
	mongoUri = "mongodb+srv://%s:%s@golang-docker-test-clus.mdqzh.gcp.mongodb.net"
)

func main() {
	r := gin.Default()
	environmentVariablesError := godotenv.Load()
	if environmentVariablesError != nil {
		panic(environmentVariablesError)
	}

	mongoAdmin := os.Getenv("MONGO_USER")
	mongoPassword := os.Getenv("MONGO_PASSWORD")

	log.Println("Mongo Admin", mongoAdmin)
	log.Println("Mongo Password", mongoPassword)

	r.GET("/", func(c *gin.Context) {
		c.String(200, "Hello there stranger!")
	})
	r.GET("/ping", func(c *gin.Context) {
		mongoContext, _ := context.WithTimeout(context.Background(), 10*time.Second)

		//mongoUri := fmt.Sprintf("mongodb+srv://%s:%s@golang-docker-test-clus.mdqzh.gcp.mongodb.net",
		//	mongoAdmin, mongoPassword)

		mongoUri := "mongodb://mongo"

		client, clientError := mongo.Connect(mongoContext, options.Client().ApplyURI(mongoUri))
		if clientError != nil {
			panic(clientError)
		}
		if pingError := client.Ping(mongoContext, readpref.Primary()); pingError != nil {
			c.JSON(http.StatusInternalServerError, bson.M{
				"message": "could not ping database",
			})
			log.Println(pingError.Error())
			return
		}
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
	r.POST("/fruits", func(c *gin.Context) {
		mongoContext, _ := context.WithTimeout(context.Background(), 10*time.Second)

		//mongoUri := fmt.Sprintf("mongodb+srv://%s:%s@golang-docker-test-clus.mdqzh.gcp.mongodb.net",
		//	mongoAdmin, mongoPassword)

		mongoUri := "mongodb://mongo"

		client, clientError := mongo.Connect(mongoContext, options.Client().ApplyURI(mongoUri))
		if clientError != nil {
			panic(clientError)
		}
		if pingError := client.Ping(mongoContext, readpref.Primary()); pingError != nil {
			c.JSON(http.StatusInternalServerError, bson.M{
				"message": "could not ping database",
			})
			log.Println(pingError.Error())
			return
		}

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
		result, insertionError := fruitsCollection.InsertMany(mongoContext, values)
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
