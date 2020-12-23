package main

import (
	"fmt"
	"time"
)

// Atom model
type Atom struct {
	Title   string  `xml:"title" json:"title"`
	Link    Link    `xml:"link" json:"link"`
	Summary string  `xml:"summary" json:"summary"`
	Entries []Entry `xml:"entry" json:"entries"`
	ID      string  `xml:"id" json:"id"`
}

// Entry model
type Entry struct {
	Title   string `xml:"title" json:"title"`
	Link    Link   `xml:"link" json:"link"`
	Summary string `xml:"summary" json:"summary"`
	Updated string `xml:"updated" json:"updated"`
	ID      string `xml:"id" json:"id"`
}

// Link model
type Link struct {
	Href string `xml:"href,attr" json:"href"`
}

func (atom *Atom) toTelegram(lastDate time.Time, chatID int) {
	for _, item := range atom.Entries {
		if item.pubTime().After(lastDate) {
			message := fmt.Sprintf("<b>%s</b>\n<a href='%s'>%s</>", atom.Title, item.Link.Href, item.Title)
			go sendMessage(MessagePayload{ChatID: chatID, Text: message, ParseMode: "HTML"})
		}
	}
}

func (item *Entry) pubTime() time.Time {
	pubDate, err := time.Parse(time.RFC3339, item.Updated)
	if err != nil {
		return time.Now()
	}
	return pubDate
}
