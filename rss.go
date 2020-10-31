package main

// Rss model
type Rss struct {
	Version     string  `xml:"version,attr" json:"version"`
	Channel     Channel `xml:"channel" json:"channel"`
	Description string  `xml:"description" json:"description"`
	Title       string  `xml:"title" json:"title"`
	Link        string  `xml:"link" json:"link"`
}

// Channel model
type Channel struct {
	Title       string `xml:"title" json:"title"`
	Link        string `xml:"link" json:"link"`
	Description string `xml:"description" json:"description"`
	Items       []Item `xml:"item" json:"items"`
}

// Item model
type Item struct {
	Title       string `xml:"title" json:"title"`
	Link        string `xml:"link" json:"link"`
	Description string `xml:"description" json:"description"`
	PubDate     string `xml:"pubDate" json:"pubdate"`
	GUID        string `xml:"guid" json:"guid"`
}

func (rss *Rss) toTelegram() []TelegramFeed {
	feeds := make([]TelegramFeed, len(rss.Channel.Items))
	for i, item := range rss.Channel.Items {
		feeds[i] = TelegramFeed{Link: item.Link, Name: rss.Channel.Title, Description: item.Description}
	}
	return feeds
}
