package main

import (
	"encoding/xml"
	"io/ioutil"
	"net/http"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

// RawFeed is a raw feed model
type RawFeed struct {
	URL   string `json:"url" firestore:"url"`
	IsRSS bool   `json:"isRSS" firestore:"isRSS"`
	Name  string `json:"name" firestore:"name"`
}

func processUsers() {
	users, err := getUsers()
	if err != nil {
		log.Error().Err(err).Msg("Error fetching users")
		return
	}
	var wg sync.WaitGroup
	for _, user := range users {
		var lastFetch time.Time
		if user.LastFetch == nil {
			lastFetch = time.Now().Add(time.Minute * -30)
			log.Info().Int("userID", user.ID).Time("lastFetch", lastFetch).
				Msg("lastFetch is NULL, setting to 30 mins ago")
		} else {
			lastFetch = *user.LastFetch
			log.Info().Int("userID", user.ID).Time("lastFetch", lastFetch).
				Msg("lastFetch is NOT NULL")
		}
		for _, f := range user.Feeds {
			wg.Add(1)
			go fetchFeed(f, lastFetch, user.ID, &wg)
		}
	}
	wg.Wait()
}

func fetchFeed(feed *PgFeed, lastUpdated time.Time, chatID int, wg *sync.WaitGroup) {
	defer updateLastFetch(chatID)
	defer wg.Done()
	resp, err := http.Get(feed.Link)
	if err != nil {
		log.Error().Err(err).Str("feedURL", feed.Link).Msg("Error fetching feed")
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error().Err(err).Str("feedURL", feed.Link).Msg("Error processing feed URL")
	}
	if feed.IsRSS {
		fd := Rss{}
		err = xml.Unmarshal(body, &fd)
		if err != nil || fd.Channel.Title == "" {
			log.Error().Str("feedURL", feed.Link).Msg("Not a valid RSS feed:")
		}
		fd.toTelegram(lastUpdated, chatID)
	} else {
		fd := Atom{}
		err = xml.Unmarshal(body, &fd)
		if err != nil || fd.Title == "" {
			log.Error().Str("feedUrl", feed.Link).Msg("Not a valid ATOM feed")
		}
		fd.toTelegram(lastUpdated, chatID)
	}
}
