package main

import (
	"context"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/labstack/echo"
)

// Rss model
type Rss struct {
	Channel Channel `xml:"channel" json:"channel"`
}

// Channel model
type Channel struct {
	Title       string `xml:"title" json:"title"`
	Description string `xml:"description" json:"description"`
	Items       []Item `xml:"item" json:"items"`
}

// Item model
type Item struct {
	Title   string `xml:"title" json:"title"`
	Link    string `xml:"link" json:"link"`
	PubDate string `xml:"pubDate" json:"pubdate"`
}

func (rss *Rss) toTelegram(fire *firestore.Client, c echo.Context) (feeds []TelegramFeed) {
	fire.Collection("userFeeds").Doc("lastFetch").Set(
		context.Background(), map[string]time.Time{"RSS": time.Now()},
	)
	for _, item := range rss.Channel.Items {
		if item.pubTime().After(time.Now()) {
			feeds = append(feeds, TelegramFeed{Link: item.Link, Name: rss.Channel.Title, Description: item.Title})
		}
	}
	if _, err := fire.Collection("userFeeds").Doc("lastFetch").Set(
		context.Background(), map[string]time.Time{"RSS": time.Now()},
	); err != nil {
		c.Logger().Error("Error saving lastFetch", err)
	}
	return feeds
}

func (item *Item) pubTime() time.Time {
	pubDate, err := time.Parse(time.RFC822Z, item.PubDate)
	if err != nil {
		return time.Now()
	}
	return pubDate
}
