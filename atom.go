package main

// Atom model
type Atom struct {
	Title   string  `xml:"title" json:"title"`
	Link    Link    `xml:"href,a" json:"link"`
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

func (atom *Atom) toTelegram() []TelegramFeed {
	feeds := make([]TelegramFeed, len(atom.Entries))
	for i, item := range atom.Entries {
		feeds[i] = TelegramFeed{Link: item.Link.Href, Name: atom.Title, Description: item.Summary}
	}
	return feeds
}
