package main

import (
	"context"
	"log"
	"os"
	"time"

	firestore "cloud.google.com/go/firestore"
	firebase "firebase.google.com/go"
)

func initClient() *firestore.Client {
	ctx := context.Background()
	conf := &firebase.Config{ProjectID: os.Getenv("PROJECT_ID")}
	app, err := firebase.NewApp(ctx, conf)
	if err != nil {
		log.Fatalln(err)
	}

	client, err := app.Firestore(ctx)
	if err != nil {
		log.Fatalln(err)
	}
	return client
}

// FireFeed is a firestore feed model
type FireFeed struct {
	URL   string `json:"url" firestore:"url"`
	IsRSS bool   `json:"isRSS" firestore:"isRSS"`
}

// LastFetch is a firestore last fetch model
type LastFetch struct {
	RSS  time.Time `firestore:"rss"`
	Atom time.Time `firestore:"atom"`
}

// FireUser model
type FireUser struct {
	UserName string   `firestore:"userName,omitempty"`
	ChatID   string   `firestore:"chatID,omitempty"`
	Feeds    []string `firestore:"feeds,omitempty"`
	Atoms    []string `firestore:"atoms,omitempty"`
}
