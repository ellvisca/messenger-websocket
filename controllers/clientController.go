package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/ellvisca/messenger-websocket/models"
	u "github.com/ellvisca/messenger-websocket/utils"
)

// Create client controller
var CreateClient = func(w http.ResponseWriter, r *http.Request) {
	client := &models.Client{}
	err := json.NewDecoder(r.Body).Decode(client)
	if err != nil {
		u.Respond(w, u.Message(false, "Invalid request"))
		return
	}

	resp := client.Create()
	u.Respond(w, resp)
}

// Login controller
var ClientLogin = func(w http.ResponseWriter, r *http.Request) {
	client := &models.Client{}
	err := json.NewDecoder(r.Body).Decode(client)
	if err != nil {
		u.Respond(w, u.Message(false, "Invalid request"))
		return
	}

	resp := models.Login(client.Username, client.Password)
	u.Respond(w, resp)
}
