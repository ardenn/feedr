package main

import (
	"fmt"
	"time"

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

func (rss *Rss) toTelegram(c echo.Context, lastDate time.Time, chatID int) {
	for _, item := range rss.Channel.Items {
		if item.pubTime().After(lastDate) {
			message := fmt.Sprintf("<b>%s</b>\n<a href='%s'>%s</>", rss.Channel.Title, item.Link, item.Title)
			sendMessage(MessagePayload{ChatID: chatID, Text: message, ParseMode: "HTML"}, c)
		}
	}
}

func (item *Item) pubTime() time.Time {
	pubDate, err := time.Parse(time.RFC1123Z, item.PubDate)
	if err != nil {
		return time.Now()
	}
	return pubDate
}
