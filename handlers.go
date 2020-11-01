package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/labstack/echo"
)

func startHandler(c echo.Context) error {
	return c.JSON(http.StatusAccepted, `{"message":"success"}`)
}
func helpHandler(c echo.Context) error {
	return c.JSON(http.StatusAccepted, `{"message":"success"}`)
}
func addHandler(c echo.Context) error {
	return c.JSON(http.StatusAccepted, `{"message":"success"}`)
}
func listHandler(c echo.Context) error {
	return c.JSON(http.StatusAccepted, `{"message":"success"}`)
}
func removeHandler(c echo.Context) error {
	return c.JSON(http.StatusAccepted, `{"message":"success"}`)
}
func commandHandler(c echo.Context) error {
	defer c.Request().Body.Close()
	update := Update{}
	body, err := ioutil.ReadAll(c.Request().Body)
	if err != nil {
		c.Logger().Error("Error reading request body", err)
	}
	json.Unmarshal(body, &update)
	switch update.Message.Text {
	case "/start":
		return startHandler(c)
	case "/help":
		return helpHandler(c)
	case "/list":
		return listHandler(c)
	case "/add":
		return addHandler(c)
	default:
		go sendMessage(MessagePayload{ChatID: update.Message.Chat.ChatID, Text: "Oops! That's an unknown command"}, c)
	}
	return c.JSON(http.StatusAccepted, `{"message":"success"}`)
}
