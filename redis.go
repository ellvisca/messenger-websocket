package main

import (
	"log"

	"github.com/gomodule/redigo/redis"
	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
)

// Channel name for redis
const Channel = "chat"

// redisReceiver receives messages from Redis and broadcasts them to all registered websocket connections.
type redisReceiver struct {
	pool             *redis.Pool
	messages         chan []byte
	newConnection    chan *websocket.Conn
	removeConnection chan *websocket.Conn
}

// Initiate new redisReceiver
func newRedisReceiver(pool *redis.Pool) redisReceiver {
	return redisReceiver{
		pool:             pool,
		messages:         make(chan []byte, 256),
		newConnection:    make(chan *websocket.Conn),
		removeConnection: make(chan *websocket.Conn),
	}
}

// Run redisReceiver pubsub messages
func (rr *redisReceiver) run() error {
	conn := rr.pool.Get()
	defer conn.Close()
	psc := redis.PubSubConn{Conn: conn}
	psc.Subscribe(Channel)
	go rr.connHandler()
	for {
		switch v := psc.Receive().(type) {
		case redis.Message:
			if _, err := validateMessage(v.Data); err != nil {
				log.Println("Error unmarshalling message from Redis")
				continue
			}
			rr.broadcast(v.Data)
		case redis.Subscription:
			log.Printf("%s: %s %d\n", v.Channel, v.Kind, v.Count)
		case error:
			return errors.Wrap(v, "Error while subscribed to Redis channel")
		default:
			log.Println("Unknown Redis receive")
		}
	}
}

// Connection handler for redisReceiver
func (rr *redisReceiver) connHandler() {
	conns := make([]*websocket.Conn, 0)
	for {
		select {
		case msg := <-rr.messages:
			for _, conn := range conns {
				err := conn.WriteMessage(websocket.TextMessage, msg)
				if err != nil {
					log.Println("Error writing message to connection")
				}
			}
		case conn := <-rr.newConnection:
			conns = append(conns, conn)
		case conn := <-rr.removeConnection:
			conns = removeConn(conns, conn)
		}
	}
}

// Remove connection function for handler
func removeConn(conns []*websocket.Conn, conn *websocket.Conn) []*websocket.Conn {
	var i int
	var exist bool
	for i = 0; i < len(conns); i++ {
		if conns[i] == conn {
			exist = true
			break
		}
	}

	if !exist {
		log.Println("Connection does not exist")
	}

	copy(conns[i:], conns[i+1:])
	conns[len(conns)-1] = nil
	return conns[:len(conns)-1]
}

func (rr *redisReceiver) connect(conn *websocket.Conn) {
	rr.newConnection <- conn
}

func (rr *redisReceiver) disconnect(conn *websocket.Conn) {
	rr.removeConnection <- conn
}

func (rr *redisReceiver) broadcast(msg []byte) {
	rr.messages <- msg
}

// redisWriter publishes messages to the Redis CHANNEL
type redisWriter struct {
	pool     *redis.Pool
	messages chan []byte
}

// Initiate new redisWriter
func newRedisWriter(pool *redis.Pool) redisWriter {
	return redisWriter{
		pool:     pool,
		messages: make(chan []byte, 256),
	}
}

// Run rediswriter
func (rw *redisWriter) run() error {
	conn := rw.pool.Get()
	defer conn.Close()

	for data := range rw.messages {
		if err := writeToRedis(conn, data); err != nil {
			rw.publish(data)
			return err
		}
	}
	return nil
}

func writeToRedis(conn redis.Conn, data []byte) error {
	if err := conn.Send("PUBLISH", Channel, data); err != nil {
		return errors.Wrap(err, "Unable to publish message to Redis")
	}
	if err := conn.Flush(); err != nil {
		return errors.Wrap(err, "Unable to flush published message to Redis")
	}
	return nil
}

func (rw *redisWriter) publish(data []byte) {
	rw.messages <- data
}
