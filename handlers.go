package main

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"cloud.google.com/go/firestore"
	"github.com/labstack/echo"
)

// IsURL validates a URL string
func IsURL(str string) bool {
	u, err := url.Parse(str)
	return err == nil && u.Host != "" && u.Scheme != ""
}

func processNewURL(url string, c echo.Context) (*FireFeed, error) {
	resp, err := http.Get(url)
	if err != nil {
		c.Logger().Errorf("Error processing new URL %s", err)
		return nil, err
	}
	defer resp.Body.Close()
	if !strings.Contains(resp.Header.Get("Content-Type"), "xml") {
		c.Logger().Errorf("Error processing new URL %s", err)
		return nil, errors.New("Oops! we couldn't get a valid feed from that URL")
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		c.Logger().Errorf("Error processing new URL %s", err)
		return nil, err
	}
	rss := Rss{}
	err = xml.Unmarshal(body, &rss)
	if err != nil || rss.Channel.Title == "" {
		c.Logger().Errorf("Not a valid RSS feed: %s", url)
		atom := Atom{}
		err = xml.Unmarshal(body, &atom)
		if err != nil || atom.Title == "" {
			c.Logger().Errorf("Not a valid ATOM feed: %s", url)
			return nil, errors.New("Oops! we couldn't get a valid feed from that URL")
		}
		return &FireFeed{URL: url, IsRSS: false}, nil
	}
	return &FireFeed{URL: url, IsRSS: true}, nil

}

func startHandler(c echo.Context, update *Update, fire *firestore.Client) error {
	user := FireUser{UserName: update.Message.From.Username, ChatID: strconv.Itoa(update.Message.Chat.ChatID)}
	_, err := fire.Collection("userFeeds").Doc(
		strconv.Itoa(update.Message.From.UserID),
	).Set(c.Request().Context(), &user)
	if err != nil {
		c.Logger().Error("Firestore error", err)
		sendMessage(MessagePayload{ChatID: update.Message.Chat.ChatID, Text: "Oops! An error occured"}, c)
		return c.JSON(http.StatusAccepted, `{"message":"success"}`)
	}
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
	/add <feed url> - Subscribe to a new feed
	/remove <feed url> - Unsubsribe from a feed
	/clear - Clear all feeds and reset account
	`
	sendMessage(MessagePayload{ChatID: update.Message.Chat.ChatID, Text: message}, c)
	return c.JSON(http.StatusAccepted, `{"message":"success"}`)
}
func addHandler(c echo.Context, update *Update, fire *firestore.Client) error {
	raw := strings.Split(update.Message.Text, "/add ")
	var rawURL string = raw[1]
	if !strings.HasPrefix(rawURL, "http") {
		rawURL = "https://" + rawURL
	}
	if _, err := url.Parse(rawURL); err != nil {
		sendMessage(MessagePayload{ChatID: update.Message.Chat.ChatID, Text: "Oops! That was an invalid URL"}, c)
		return c.JSON(http.StatusAccepted, `{"message":"success"}`)
	}
	fireFeed, err := processNewURL(rawURL, c)
	if err != nil {
		sendMessage(MessagePayload{ChatID: update.Message.Chat.ChatID, Text: err.Error()}, c)
		return c.JSON(http.StatusAccepted, `{"message":"success"}`)
	}
	user := FireUser{}
	doc, err := fire.Collection("userFeeds").Doc(strconv.Itoa(update.Message.From.UserID)).Get(c.Request().Context())
	if err != nil {
		c.Logger().Error("Firestore error", err)
		sendMessage(MessagePayload{ChatID: update.Message.Chat.ChatID, Text: "Oops! An error occured when saving feed"}, c)
		return c.JSON(http.StatusAccepted, `{"message":"success"}`)
	}
	if err = doc.DataTo(&user); err != nil {
		c.Logger().Error("Firestore error", err)
		sendMessage(MessagePayload{ChatID: update.Message.Chat.ChatID, Text: "Oops! An error occured when saving feed"}, c)
		return c.JSON(http.StatusAccepted, `{"message":"success"}`)
	}
	if fireFeed.IsRSS {
		user.Feeds = append(user.Feeds, fireFeed.URL)
	} else {
		user.Atoms = append(user.Atoms, fireFeed.URL)
	}
	user.ChatID = strconv.Itoa(update.Message.Chat.ChatID)
	user.UserName = update.Message.From.Username
	_, err = fire.Collection("userFeeds").Doc(
		strconv.Itoa(update.Message.From.UserID),
	).Set(c.Request().Context(), map[string]interface{}{
		"feeds": user.Feeds,
		"atoms": user.Atoms,
	}, firestore.MergeAll)
	if err != nil {
		c.Logger().Error("Firestore error", err)
		sendMessage(MessagePayload{ChatID: update.Message.Chat.ChatID, Text: "Oops! An error occured when saving feed"}, c)
		return c.JSON(http.StatusAccepted, `{"message":"success"}`)
	}
	sendMessage(MessagePayload{ChatID: update.Message.Chat.ChatID, Text: "Success! Feed has been added"}, c)
	return c.JSON(http.StatusAccepted, `{"message":"success"}`)
}
func listHandler(c echo.Context, update *Update, fire *firestore.Client) error {
	message := "Your feeds:\n"
	user := FireUser{}
	doc, err := fire.Collection("userFeeds").Doc(
		strconv.Itoa(update.Message.From.UserID),
	).Get(c.Request().Context())
	if err != nil {
		sendMessage(MessagePayload{ChatID: update.Message.Chat.ChatID, Text: "Oops! An error occurred when fetching feeds"}, c)
		c.Logger().Errorf("Error reading firestore feed list", err)
	}
	if err = doc.DataTo(&user); err != nil {
		sendMessage(MessagePayload{ChatID: update.Message.Chat.ChatID, Text: "Oops! An error occurred when fetching feeds"}, c)
		c.Logger().Errorf("Error reading firestore feed list", err)
	}
	feeds := append(user.Feeds, user.Atoms...)
	if feeds == nil {
		message = "You haven't added any feeds."
	} else {
		for _, link := range feeds {
			val, err := url.Parse(link)
			if err != nil {
				continue
			}
			message += fmt.Sprintln("-", val.Host)
		}
	}
	sendMessage(MessagePayload{ChatID: update.Message.Chat.ChatID, Text: message}, c)
	return c.JSON(http.StatusAccepted, `{"message":"success"}`)
}
func removeHandler(c echo.Context, update *Update, fire *firestore.Client) error {
	return c.JSON(http.StatusAccepted, `{"message":"success"}`)
}
func commandHandler(c echo.Context) error {
	cc := c.(*CustomContext)
	defer c.Request().Body.Close()
	update := Update{}
	body, err := ioutil.ReadAll(c.Request().Body)
	if err != nil {
		c.Logger().Errorf("Error reading request body", err)
	}
	json.Unmarshal(body, &update)
	switch {
	case update.Message.Text == "/start":
		return startHandler(c, &update, cc.fire)
	case update.Message.Text == "/help":
		return helpHandler(c, &update)
	case update.Message.Text == "/list":
		return listHandler(c, &update, cc.fire)
	case strings.HasPrefix(update.Message.Text, "/add "):
		fmt.Println(update.Message.Text)
		return addHandler(c, &update, cc.fire)
	default:
		sendMessage(MessagePayload{ChatID: update.Message.Chat.ChatID, Text: "Oops! That's an unknown command"}, c)
	}
	return c.JSON(http.StatusAccepted, `{"message":"success"}`)
}
func crawlHandler(c echo.Context) error {
	cc := c.(*CustomContext)
	fetchUsers(cc.fire, c)
	return c.JSON(http.StatusAccepted, `{"message":"success"}`)
}
