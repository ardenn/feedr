package main

import (
	"encoding/xml"
	"io/ioutil"
	"net/http"
	"strconv"
	"sync"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/labstack/echo"
	"google.golang.org/api/iterator"
)

func fetchUsers(fire *firestore.Client, c echo.Context) {
	iter := fire.Collection("userFeeds").Documents(c.Request().Context())
	var wg sync.WaitGroup
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			c.Logger().Errorf("Error reading crawlFeeds", err)
		}
		user := FireUser{}
		if err := doc.DataTo(&user); err != nil {
			c.Logger().Errorf("Error reading firestore users list", err)
		}
		lastFetch := time.Now().Add(time.Minute * -30)
		chatID, _ := strconv.ParseInt(user.ChatID, 0, 64)
		for _, f := range user.Feeds {
			wg.Add(1)
			go fetchFeed(f, true, c, lastFetch, int(chatID), &wg)
		}
		for _, f := range user.Atoms {
			wg.Add(1)
			go fetchFeed(f, false, c, lastFetch, int(chatID), &wg)
		}
	}
	wg.Wait()
}

func fetchFeed(str string, isRSS bool, c echo.Context, lastUpdated time.Time, chatID int, wg *sync.WaitGroup) {
	resp, err := http.Get(str)
	if err != nil {
		c.Logger().Error("Error fetching feed", str, err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		c.Logger().Errorf("Error processing feed URL", str, err)
	}
	if isRSS {
		feed := Rss{}
		err = xml.Unmarshal(body, &feed)
		if err != nil || feed.Channel.Title == "" {
			c.Logger().Error("Not a valid RSS feed:", str)
		}
		feed.toTelegram(c, lastUpdated, chatID)
	} else {
		feed := Atom{}
		err = xml.Unmarshal(body, &feed)
		if err != nil || feed.Title == "" {
			c.Logger().Errorf("Not a valid ATOM feed: %s", str)
		}
		feed.toTelegram(c, lastUpdated, chatID)
	}
	wg.Done()
}
