package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/ellvisca/messenger-websocket/controllers"
	u "github.com/ellvisca/messenger-websocket/utils"
	"github.com/gomodule/redigo/redis"
	"github.com/joho/godotenv"
)

var (
	rr redisReceiver
	rw redisWriter
)

func newPool(addr string) *redis.Pool {
	return &redis.Pool{
		MaxIdle:     3,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp",
				addr,
				redis.DialPassword(os.Getenv("REDIS_PASSWORD")))
			if err != nil {
				return nil, err
			}
			return c, nil
		},
	}
}

func Home(w http.ResponseWriter, r *http.Request) {
	u.Respond(w, u.Message(true, "Welcome to API"))
}

func main() {
	err := godotenv.Load()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		log.Println(`$REDIS_URL must be set`)
	}

	redisPool := newPool(redisURL)
	rr = newRedisReceiver(redisPool)
	rw = newRedisWriter(redisPool)

	go func() {
		for {
			err = rr.run()
			if err == nil {
				break
			}
		}
	}()

	go func() {
		for {
			err = rw.run()
			if err == nil {
				break
			}
		}
	}()

	// Serve home
	http.HandleFunc("/", Home)

	// Serve ws
	http.HandleFunc("/ws", serveWs)

	// Create client
	http.HandleFunc("/api/v1/client", controllers.CreateClient)

	log.Println("Listening on port", port)
	log.Println(http.ListenAndServe(":"+port, nil))
}
