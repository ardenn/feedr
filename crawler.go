package main

import (
	"context"
	"encoding/xml"
	"io/ioutil"
	"net/http"
	"strconv"
	"sync"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/rs/zerolog/log"
	"google.golang.org/api/iterator"
)

func fetchUsers(ctx context.Context, fire *firestore.Client) {
	iter := fire.Collection("userFeeds").Documents(ctx)
	var wg sync.WaitGroup
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Error().Str("error", err.Error()).Msg("Error reading crawl Feeds")
		}
		user := FireUser{}
		if err := doc.DataTo(&user); err != nil {
			log.Error().Str("error", err.Error()).Msg("Error reading firestore users list")
		}
		lastFetch := time.Now().Add(time.Minute * -30)
		chatID, _ := strconv.ParseInt(user.ChatID, 0, 64)
		for _, f := range user.Feeds {
			wg.Add(1)
			go fetchFeed(f, true, lastFetch, int(chatID), &wg)
		}
		for _, f := range user.Atoms {
			wg.Add(1)
			go fetchFeed(f, false, lastFetch, int(chatID), &wg)
		}
	}
	wg.Wait()
}

func fetchFeed(str string, isRSS bool, lastUpdated time.Time, chatID int, wg *sync.WaitGroup) {
	resp, err := http.Get(str)
	if err != nil {
		log.Error().Str("error", err.Error()).Str("feedUrl", str).Msg("Error fetching feed")
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error().Str("error", err.Error()).Str("feedUrl", str).Msg("Error processing feed URL")
	}
	if isRSS {
		feed := Rss{}
		err = xml.Unmarshal(body, &feed)
		if err != nil || feed.Channel.Title == "" {
			log.Error().Str("feedUrl", str).Msg("Not a valid RSS feed:")
		}
		feed.toTelegram(lastUpdated, chatID)
	} else {
		feed := Atom{}
		err = xml.Unmarshal(body, &feed)
		if err != nil || feed.Title == "" {
			log.Error().Str("feedUrl", str).Msg("Not a valid ATOM feed")
		}
		feed.toTelegram(lastUpdated, chatID)
	}
	wg.Done()
}
