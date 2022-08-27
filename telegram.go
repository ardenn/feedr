package main

import (
	"bytes"
	"encoding/json"
	"io"
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

// TelegramChat model
type TelegramChat struct {
	ChatID int `json:"id"`
}

// TelegramUser model
type TelegramUser struct {
	UserID   int    `json:"id"`
	IsBot    bool   `json:"is_bot"`
	Username string `json:"username"`
}

// TelegramMessage is Telegram TelegramMessage Object
type TelegramMessage struct {
	MessageID int          `json:"message_id"`
	Chat      TelegramChat `json:"chat"`
	Text      string       `json:"text"`
	From      TelegramUser `json:"from"`
}

// TelegramUpdate is a Telegram TelegramUpdate Object
type TelegramUpdate struct {
	UpdateID int             `json:"update_id"`
	Message  TelegramMessage `json:"message"`
}

// TelegramMessagePayload defines the request payload to sendMessage on Telegram
type TelegramMessagePayload struct {
	ChatID    int    `json:"chat_id"`
	Text      string `json:"text"`
	ParseMode string `json:"parse_mode,omitempty"`
}

// TelegramResponse model
type TelegramResponse struct {
	OK          bool   `json:"ok"`
	Description string `json:"description"`
}

func sendMessage(message TelegramMessagePayload) {
	apiURL := "https://api.telegram.org/bot" + os.Getenv("BOT_TOKEN") + "/"
	payload, _ := json.Marshal(message)
	resp, err := http.Post(apiURL+"sendMessage", "application/json", bytes.NewBuffer(payload))
	if err != nil {
		log.Error().Err(err).Str("body", string(payload)).Msg("Error sending message to Telegram")
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error().Err(err).Msg("Error reading Telegram response")
	}
	response := TelegramResponse{}
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
