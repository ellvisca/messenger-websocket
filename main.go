package main

import (
	"log"
	"net/http"
	"os"
	"time"

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

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
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
	http.Handle("/", http.FileServer(http.Dir("./public")))

	// Serve ws
	http.HandleFunc("/ws", serveWs)

	log.Println("Listening on port", port)
	log.Println(http.ListenAndServe(":"+port, nil))
}
