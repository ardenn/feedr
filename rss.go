package main

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

func (rss *Rss) toTelegram() []TelegramFeed {
	feeds := make([]TelegramFeed, len(rss.Channel.Items))
	for i, item := range rss.Channel.Items {
		feeds[i] = TelegramFeed{Link: item.Link, Name: rss.Channel.Title, Description: item.Title}
	}
	return feeds
}
