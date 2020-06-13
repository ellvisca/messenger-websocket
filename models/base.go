package models

import (
	"context"
	"log"
	"os"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var db *mongo.Database

func init() {
	// Load .env
	godotenv.Load()

	// Connect to MongoDB
	clientOptions := options.Client().ApplyURI(os.Getenv("MONGO_URL"))

	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Println(err)
	}

	log.Println("Connected to MongoDB!")
	db = client.Database("messenger")
}

func GetDB() *mongo.Database {
	return db
}
