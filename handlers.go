package main

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/rs/zerolog/log"
)

// IsURL validates a URL string
func IsURL(str string) bool {
	u, err := url.Parse(str)
	return err == nil && u.Host != "" && u.Scheme != ""
}

func processNewURL(url string) (*RawFeed, error) {
	resp, err := http.Get(url)
	if err != nil {
		log.Error().Err(err).Str("feedUrl", url).Msg("Error fetching new URL")
		return nil, err
	}
	defer resp.Body.Close()
	if !strings.Contains(resp.Header.Get("Content-Type"), "xml") {
		log.Error().Str("feedUrl", url).Msg("URL response not valid xml")
		return nil, errors.New("Oops! we couldn't get a valid feed from that URL")
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error().Err(err).Str("feedUrl", url).Msg("Error processing new URL")
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
		return &RawFeed{URL: url, IsRSS: false, Name: atom.Title}, nil
	}
	return &RawFeed{URL: url, IsRSS: true, Name: rss.Channel.Title}, nil

}

func startHandler(update *Update) {
	userID, _ := addUser(&update.Message)
	if userID == 0 {
		go sendMessage(MessagePayload{ChatID: update.Message.Chat.ChatID, Text: "Oops! Something went wrong!"})
		return
	}
	message := `
	Welcome to Feedr.

	Add feeds (atom/rss) and we'll subscribe and ping you whenever there's an update.
	`
	go sendMessage(MessagePayload{ChatID: update.Message.Chat.ChatID, Text: message})
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
	go sendMessage(MessagePayload{ChatID: update.Message.Chat.ChatID, Text: message})
	return
}

func addHandler(update *Update) {
	raw := strings.Split(update.Message.Text, "/add ")
	var rawURL string = raw[1]
	if !strings.HasPrefix(rawURL, "http") {
		rawURL = "https://" + rawURL
	}
	if _, err := url.Parse(rawURL); err != nil {
		log.Info().Str("feedURL", rawURL).Msg("Invalid URL")
		go sendMessage(MessagePayload{ChatID: update.Message.Chat.ChatID, Text: "Oops! That was an invalid URL"})
		return
	}
	rawFeed, err := processNewURL(rawURL)
	if err != nil {
		go sendMessage(MessagePayload{ChatID: update.Message.Chat.ChatID, Text: err.Error()})
		return
	}
	if addFeed(rawFeed, &update.Message) {
		go sendMessage(MessagePayload{ChatID: update.Message.Chat.ChatID, Text: "Success! Feed has been added"})
		return
	}
	go sendMessage(MessagePayload{ChatID: update.Message.Chat.ChatID, Text: "Oops! An error occured when saving feed"})
	return
}

func listHandler(update *Update) {
	message := "Your feeds:\n"
	feeds, err := getUserFeeds(update.Message.From.UserID)
	if err != nil {
		log.Error().Err(err).Msg("Error reading feed list")
		go sendMessage(MessagePayload{
			ChatID: update.Message.Chat.ChatID,
			Text:   "Oops! An error occurred when fetching feeds",
		})
		return
	}

	if len(feeds) == 0 {
		message = "You haven't added any feeds."
	} else {
		for _, feed := range feeds {
			val, err := url.Parse(feed.Link)
			if err != nil {
				continue
			}
			message += fmt.Sprintln("-", val.Host)
		}
	}
	go sendMessage(MessagePayload{ChatID: update.Message.Chat.ChatID, Text: message})
	return
}

func removeHandler(update *Update) {
	return
}

func commandHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	update := Update{}
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Error().Err(err).Msg("Error reading request body")
	}
	json.Unmarshal(body, &update)
	switch {
	case update.Message.Text == "/start":
		startHandler(&update)
	case update.Message.Text == "/help":
		helpHandler(&update)
	case update.Message.Text == "/list":
		listHandler(&update)
	case strings.HasPrefix(update.Message.Text, "/add "):
		addHandler(&update)
	default:
		log.Info().Str("command", update.Message.Text).Msg("Invalid command")
		go sendMessage(MessagePayload{ChatID: update.Message.Chat.ChatID, Text: "Oops! That's an unknown command"})
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, `{"message":"success"}`)
}

func crawlHandler(w http.ResponseWriter, r *http.Request) {
	processUsers()
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, `{"message":"success"}`)
}
