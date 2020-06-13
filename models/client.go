package models

import (
	"context"
	"log"
	"os"

	"github.com/Kamva/mgm"
	"github.com/dgrijalva/jwt-go"
	u "github.com/ellvisca/messenger-websocket/utils"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

type Client struct {
	mgm.DefaultModel `bson:",inline"`
	Username         string `json:"username"`
	Password         string `json:"password,omitempty"`
	Token            string `json:"token,omitempty"`
}

type Token struct {
	ClientId primitive.ObjectID
	jwt.StandardClaims
}

// Validate incoming register request
func (client *Client) Validate() (map[string]interface{}, bool) {
	// Check for duplicate username
	collection := GetDB().Collection("clients")
	filter := bson.M{"username": client.Username}
	err := collection.FindOne(context.TODO(), filter).Decode(&client)
	if err == nil {
		return u.Message(false, "Username already taken"), false
	}

	// Check for password length
	if len(client.Password) < 6 {
		return u.Message(false, "Password needs to be at least 6 characters"), false
	}

	// Valid response
	return u.Message(false, "Requirement passed"), true
}

// Create new client
func (client *Client) Create() map[string]interface{} {
	collection := GetDB().Collection("clients")

	// Validation
	if resp, ok := client.Validate(); !ok {
		return resp
	}

	// Hash password
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(client.Password), bcrypt.DefaultCost)
	client.Password = string(hashedPassword)

	// Create user attempt
	doc, err := collection.InsertOne(context.TODO(), client)
	if err != nil {
		return u.Message(false, "Connection error, please try again")
	}
	id := doc.InsertedID.(primitive.ObjectID)

	// Response
	filter := bson.M{"_id": id}
	collection.FindOne(context.TODO(), filter).Decode(&client)
	client.Password = ""
	resp := u.Message(true, "Successfully created client")
	resp["data"] = client
	return resp
}

// Client login
func Login(username, password string) map[string]interface{} {
	// Load .env
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	collection := GetDB().Collection("clients")
	filter := bson.M{"username": username}
	client := &Client{}

	// Log in attempt
	err = collection.FindOne(context.TODO(), filter).Decode(&client)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return u.Message(false, "Username not found")
		}
		return u.Message(false, "Connection error, please try again")
	}

	err = bcrypt.CompareHashAndPassword([]byte(client.Password), []byte(password))
	if err != nil {
		if err == bcrypt.ErrMismatchedHashAndPassword {
			return u.Message(false, "Invalid login credentials")
		}
	}

	// Token
	tk := &Token{ClientId: client.ID}
	token := jwt.NewWithClaims(jwt.GetSigningMethod("HS256"), tk)
	tokenString, _ := token.SignedString([]byte(os.Getenv("SECRET_KEY")))
	client.Token = tokenString

	// Response
	client.Password = ""
	resp := u.Message(true, "Successfully logged in")
	resp["data"] = client
	return resp
}
