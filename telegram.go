package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/rs/zerolog/log"
)

// TelegramFeed model
type TelegramFeed struct {
	Name        string `json:"name"`
	Link        string `json:"link"`
	Description string `json:"description"`
}

// Chat model
type Chat struct {
	ChatID int `json:"id"`
}

// User model
type User struct {
	UserID   int    `json:"id"`
	IsBot    bool   `json:"is_bot"`
	Username string `json:"username"`
}

// Message is Telegram Message Object
type Message struct {
	MessageID int    `json:"message_id"`
	Chat      Chat   `json:"chat"`
	Text      string `json:"text"`
	From      User   `json:"from"`
}

// Update is a Telegram Update Object
type Update struct {
	UpdateID int     `json:"update_id"`
	Message  Message `json:"message"`
}

// MessagePayload defines the request payload to sendMessage on Telegram
type MessagePayload struct {
	ChatID    int    `json:"chat_id"`
	Text      string `json:"text"`
	ParseMode string `json:"parse_mode,omitempty"`
}

// Response model
type Response struct {
	OK          bool   `json:"ok"`
	Description string `json:"description"`
}

func sendMessage(message MessagePayload) {
	apiURL := "https://api.telegram.org/bot" + os.Getenv("BOT_TOKEN") + "/"
	payload, _ := json.Marshal(message)
	resp, err := http.Post(apiURL+"sendMessage", "application/json", bytes.NewBuffer(payload))
	if err != nil {
		log.Error().Err(err).Str("body", string(payload)).Msg("Error sending message to Telegram")
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error().Err(err).Msg("Error reading Telegram response")
	}
	response := Response{}
	err = json.Unmarshal(body, &response)
	if err != nil {
		log.Error().Err(err).Msg("Error decoding Telegram response body")
	}
	if !response.OK {
		log.Info().Str("description", response.Description).Msg("Telegram request unsuccesful")
	} else {
		log.Info().Msg("Telegram message sent successfully")
	}
}
