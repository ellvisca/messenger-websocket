package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type Message struct {
	Client string `json:"client"`
	Text   string `json:"text"`
}

func validateMessage(data []byte) (Message, error) {
	var msg Message

	err := json.Unmarshal(data, &msg)
	if err != nil {
		return msg, errors.Wrap(err, "Unmarshalling error")
	}

	if msg.Client == "" || msg.Text == "" {
		{
			return msg, errors.New("Message has no client or text information")
		}
	}

	return msg, nil
}

func serveWs(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	upgrader.CheckOrigin = func(r *http.Request) bool { return true }

	// WS upgrader
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		http.Error(w, "Error upgrader", http.StatusBadRequest)
		return
	}

	// Connect to redisReceiver
	rr.connect(ws)

	// Read message
	for {
		msgtype, data, err := ws.ReadMessage()
		if err != nil {
			log.Println("Error reading websocket message")
			break
		}

		switch msgtype {
		case websocket.TextMessage:
			_, err := validateMessage(data)
			if err != nil {
				log.Println("Error validating message")
				break
			}
			rw.publish(data)
		default:
			log.Println("Unknown message")
		}
	}

	// Disconnect websocket
	rr.disconnect(ws)

	// Write message
	ws.WriteMessage(websocket.CloseMessage, []byte{})
}
