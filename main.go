package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gomodule/redigo/redis"
)

var (
	rr redisReceiver
	rw redisWriter
)

func main() {
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
			rr.broadcast(availableMessage)
			rr.run()
		}
	}()

	go func() {
		for {
			rw.run()
		}
	}()

	// Serve home
	//

	// Serve ws
	http.HandleFunc("/ws", serveWs)

	log.Println("Listening on port ", port)
	log.Println(http.ListenAndServe(":"+port, nil))
}

func newPool(addr string) *redis.Pool {
	return &redis.Pool{
		MaxIdle:     3,
		IdleTimeout: 240 * time.Second,
		Dial:        func() (redis.Conn, error) { return redis.Dial("tcp", addr) },
	}
}
