package main

import (
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"cloud.google.com/go/firestore"
	"github.com/labstack/echo"
	"google.golang.org/api/iterator"
)

// IsURL validates a URL string
func IsURL(str string) bool {
	u, err := url.Parse(str)
	return err == nil && u.Host != "" && u.Scheme != ""
}

func processNewURL(url string, c echo.Context) (*FireFeed, error) {
	resp, err := http.Get(url)
	if err != nil {
		c.Logger().Error("Error processing new URL %s", err)
		return nil, err
	}
	defer resp.Body.Close()
	if !strings.Contains(resp.Header.Get("ContentType"), "xml") {
		c.Logger().Error("Error processing new URL %s", err)
		return nil, errors.New("Oops! we couldn't get a valid feed from that URL")
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		c.Logger().Error("Error processing new URL %s", err)
		return nil, err
	}
	rss := Rss{}
	err = json.Unmarshal(body, &rss)
	if err != nil || rss.Title == "" {
		atom := Atom{}
		err = json.Unmarshal(body, &atom)
		if err != nil || atom.Title == "" {
			c.Logger().Error("Not a valid ATOM feed: %s", url)
			return nil, errors.New("Oops! we couldn't get a valid feed from that URL")
		}
		return &FireFeed{URL: url, IsRSS: false}, nil
	}
	return &FireFeed{URL: url, IsRSS: true}, nil

}

func startHandler(c echo.Context, update *Update) error {
	message := `
	Welcome to Feedr.
	Add feeds (atom/rss) and we'll subscribe and ping you whenever there's an update.
	`
	sendMessage(MessagePayload{ChatID: update.Message.Chat.ChatID, Text: message}, c)
	return c.JSON(http.StatusAccepted, `{"message":"success"}`)
}
func helpHandler(c echo.Context, update *Update) error {
	message := `
	You can control me by sending these commands:

	/help - Show this help text
	/list - List your subscribed feeds
	/add<feed url> - Subscribe to a new feed
	/remove<feed url> - Unsubsribe from a feed
	/clear - Clear all feeds and reset account
	`
	sendMessage(MessagePayload{ChatID: update.Message.Chat.ChatID, Text: message}, c)
	return c.JSON(http.StatusAccepted, `{"message":"success"}`)
}
func addHandler(c echo.Context, update *Update, fire *firestore.Client, ctx *context.Context) error {
	raw := strings.Split(update.Message.Text, "/add")
	var rawURL string = raw[1]
	if !strings.HasPrefix(rawURL, "http") {
		rawURL = "https://" + rawURL
	}
	if _, err := url.Parse(rawURL); err != nil {
		sendMessage(MessagePayload{ChatID: update.Message.Chat.ChatID, Text: "Oops! That was an invalid URL"}, c)
	} else {
		fireFeed, err := processNewURL(rawURL, c)
		if err != nil {
			sendMessage(MessagePayload{ChatID: update.Message.Chat.ChatID, Text: err.Error()}, c)
		}
		_, _, err = fire.Collection("userFeeds").Doc(strconv.Itoa(update.Message.From.UserID)).Collection("feeds").Add(*ctx, fireFeed)
		if err != nil {
			sendMessage(MessagePayload{ChatID: update.Message.Chat.ChatID, Text: "Oops! An error occured when saving feed"}, c)
		}
	}
	return c.JSON(http.StatusAccepted, `{"message":"success"}`)
}
func listHandler(c echo.Context, update *Update, fire *firestore.Client, ctx *context.Context) error {
	message := "Your feeds:\n"
	var feed FireFeed
	iter := fire.Collection("userFeeds").Doc(strconv.Itoa(update.Message.From.UserID)).Collection("feeds").Documents(*ctx)
	defer iter.Stop()
	index := 0
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			if index == 0 {
				sendMessage(MessagePayload{ChatID: update.Message.Chat.ChatID, Text: "You have 0 feeds"}, c)
			}
			break
		}
		if err != nil {
			sendMessage(MessagePayload{ChatID: update.Message.Chat.ChatID, Text: "Oops! An error occurred when fetching feeds"}, c)
			c.Logger().Error("Error reading firestore feed list", err)
		}
		if err := doc.DataTo(&feed); err != nil {
			sendMessage(MessagePayload{ChatID: update.Message.Chat.ChatID, Text: "Oops! An error occurred when fetching feeds"}, c)
			c.Logger().Error("Error reading firestore feed list", err)
		}
		message += strconv.Itoa(index+1) + ".\t" + feed.URL + "\n"
		index++
	}
	sendMessage(MessagePayload{ChatID: update.Message.Chat.ChatID, Text: message}, c)
	return c.JSON(http.StatusAccepted, `{"message":"success"}`)
}
func removeHandler(c echo.Context, update *Update, fire *firestore.Client, ctx *context.Context) error {
	return c.JSON(http.StatusAccepted, `{"message":"success"}`)
}
func commandHandler(c echo.Context) error {
	cc := c.(*CustomContext)
	defer c.Request().Body.Close()
	update := Update{}
	body, err := ioutil.ReadAll(c.Request().Body)
	if err != nil {
		c.Logger().Error("Error reading request body", err)
	}
	json.Unmarshal(body, &update)
	switch {
	case update.Message.Text == "/start":
		return startHandler(c, &update)
	case update.Message.Text == "/help":
		return helpHandler(c, &update)
	case update.Message.Text == "/list":
		return listHandler(c, &update, cc.fire, cc.ctx)
	case strings.HasPrefix(update.Message.Text, "/add"):
		return addHandler(c, &update, cc.fire, cc.ctx)
	default:
		sendMessage(MessagePayload{ChatID: update.Message.Chat.ChatID, Text: "Oops! That's an unknown command"}, c)
	}
	return c.JSON(http.StatusAccepted, `{"message":"success"}`)
}
