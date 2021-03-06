package main

import (
	"fmt"
	"net/url"
	"time"
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

func (rss *Rss) toTelegram(lastDate time.Time, chatID int, rHash string) {
	for _, item := range rss.Channel.Items {
		if item.pubTime().After(lastDate) {
			if rHash != "" {
				item.Link = fmt.Sprintf("https://t.me/iv?url=%s&rhash=%s", url.QueryEscape(item.Link), rHash)
			}
			message := fmt.Sprintf("<b>%s</b>\n<a href='%s'>%s</>", rss.Channel.Title, item.Link, item.Title)
			sendMessage(TelegramMessagePayload{ChatID: chatID, Text: message, ParseMode: "HTML"})
		}
	}
}

func (item *Item) pubTime() time.Time {
	pubDate, err := time.Parse(time.RFC1123Z, item.PubDate)
	if err != nil {
		pubDate, err = time.Parse(time.RFC3339, item.PubDate)
		if err != nil {
			return time.Now().Add(time.Hour * -24)
		}
	}
	return pubDate
}
