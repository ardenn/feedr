package main

import (
	"context"
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
	"github.com/rs/zerolog/log"
)

// IsURL validates a URL string
func IsURL(str string) bool {
	u, err := url.Parse(str)
	return err == nil && u.Host != "" && u.Scheme != ""
}

func processNewURL(ctx context.Context, url string) (*FireFeed, error) {
	resp, err := http.Get(url)
	if err != nil {
		log.Error().Str("error", err.Error()).Str("feedUrl", url).Msg("Error fetching new URL")
		return nil, err
	}
	defer resp.Body.Close()
	if !strings.Contains(resp.Header.Get("Content-Type"), "xml") {
		log.Error().Str("error", err.Error()).Str("feedUrl", url).Msg("URL response not valid xml")
		return nil, errors.New("Oops! we couldn't get a valid feed from that URL")
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error().Str("error", err.Error()).Str("feedUrl", url).Msg("Error processing new URL")
		return nil, err
	}
	rss := Rss{}
	err = xml.Unmarshal(body, &rss)
	if err != nil || rss.Channel.Title == "" {
		log.Error().Str("feedUrl", url).Msg("Not a valid RSS feed")
		atom := Atom{}
		err = xml.Unmarshal(body, &atom)
		if err != nil || atom.Title == "" {
			log.Error().Str("feedUrl", url).Msg("Not a valid ATOM feed")
			return nil, errors.New("Oops! we couldn't get a valid feed from that URL")
		}
		return &FireFeed{URL: url, IsRSS: false}, nil
	}
	return &FireFeed{URL: url, IsRSS: true}, nil

}

func startHandler(ctx context.Context, update *Update, fire *firestore.Client) {
	user := FireUser{UserName: update.Message.From.Username, ChatID: strconv.Itoa(update.Message.Chat.ChatID)}
	_, err := fire.Collection("userFeeds").Doc(
		strconv.Itoa(update.Message.From.UserID),
	).Set(ctx, &user)
	if err != nil {
		log.Error().Str("error", err.Error()).Msg("Firestore error setting /start user")
		sendMessage(MessagePayload{ChatID: update.Message.Chat.ChatID, Text: "Oops! An error occured"})
		return
	}
	message := `
	Welcome to Feedr.

	Add feeds (atom/rss) and we'll subscribe and ping you whenever there's an update.
	`
	sendMessage(MessagePayload{ChatID: update.Message.Chat.ChatID, Text: message})
	return
}
func helpHandler(update *Update) {
	message := `
	You can control me by sending these commands:

	/help - Show this help text
	/list - List your subscribed feeds
	/add <feed url> - Subscribe to a new feed
	/remove <feed url> - Unsubsribe from a feed
	/clear - Clear all feeds and reset account
	`
	sendMessage(MessagePayload{ChatID: update.Message.Chat.ChatID, Text: message})
	return
}
func addHandler(ctx context.Context, update *Update, fire *firestore.Client) {
	raw := strings.Split(update.Message.Text, "/add ")
	var rawURL string = raw[1]
	if !strings.HasPrefix(rawURL, "http") {
		rawURL = "https://" + rawURL
	}
	if _, err := url.Parse(rawURL); err != nil {
		log.Printf("Invalid URL %s", rawURL)
		sendMessage(MessagePayload{ChatID: update.Message.Chat.ChatID, Text: "Oops! That was an invalid URL"})
		return
	}
	fireFeed, err := processNewURL(ctx, rawURL)
	if err != nil {
		sendMessage(MessagePayload{ChatID: update.Message.Chat.ChatID, Text: err.Error()})
		return
	}
	user := FireUser{}
	doc, err := fire.Collection("userFeeds").Doc(strconv.Itoa(update.Message.From.UserID)).Get(ctx)
	if err != nil {
		log.Error().Str("error", err.Error()).Msg("Firestore error")
		sendMessage(MessagePayload{ChatID: update.Message.Chat.ChatID, Text: "Oops! An error occured when saving feed"})
		return
	}
	if err = doc.DataTo(&user); err != nil {
		log.Error().Str("error", err.Error()).Msg("Firestore error")
		sendMessage(MessagePayload{ChatID: update.Message.Chat.ChatID, Text: "Oops! An error occured when saving feed"})
		return
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
	).Set(ctx, map[string]interface{}{
		"feeds": user.Feeds,
		"atoms": user.Atoms,
	}, firestore.MergeAll)
	if err != nil {
		log.Error().Str("error", err.Error()).Msg("Firestore error")
		sendMessage(MessagePayload{ChatID: update.Message.Chat.ChatID, Text: "Oops! An error occured when saving feed"})
		return
	}
	sendMessage(MessagePayload{ChatID: update.Message.Chat.ChatID, Text: "Success! Feed has been added"})
	return
}
func listHandler(ctx context.Context, update *Update, fire *firestore.Client) {
	message := "Your feeds:\n"
	user := FireUser{}
	doc, err := fire.Collection("userFeeds").Doc(
		strconv.Itoa(update.Message.From.UserID),
	).Get(ctx)
	if err != nil {
		log.Error().Str("error", err.Error()).Msg("Error reading firestore feed list")
		sendMessage(MessagePayload{ChatID: update.Message.Chat.ChatID, Text: "Oops! An error occurred when fetching feeds"})
	}
	if err = doc.DataTo(&user); err != nil {
		log.Error().Str("error", err.Error()).Msg("Error reading firestore feed list")
		sendMessage(MessagePayload{ChatID: update.Message.Chat.ChatID, Text: "Oops! An error occurred when fetching feeds"})
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
	sendMessage(MessagePayload{ChatID: update.Message.Chat.ChatID, Text: message})
	return
}
func removeHandler(c context.Context, update *Update, fire *firestore.Client) {
	return
}
func commandHandler(w http.ResponseWriter, r *http.Request) {
	fire := FireContext(r.Context())
	defer r.Body.Close()
	update := Update{}
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Error().Str("error", err.Error()).Msg("Error reading request body")
	}
	json.Unmarshal(body, &update)
	switch {
	case update.Message.Text == "/start":
		startHandler(r.Context(), &update, fire)
	case update.Message.Text == "/help":
		helpHandler(&update)
	case update.Message.Text == "/list":
		listHandler(r.Context(), &update, fire)
	case strings.HasPrefix(update.Message.Text, "/add "):
		addHandler(r.Context(), &update, fire)
	default:
		log.Info().Str("command", update.Message.Text).Msg("Invalid command")
		sendMessage(MessagePayload{ChatID: update.Message.Chat.ChatID, Text: "Oops! That's an unknown command"})
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, `{"message":"success"}`)
}

func crawlHandler(w http.ResponseWriter, r *http.Request) {
	fire := FireContext(r.Context())
	fetchUsers(r.Context(), fire)
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, `{"message":"success"}`)
}
