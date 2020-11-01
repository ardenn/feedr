package main

import (
	"context"
	"log"
	"os"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go"
)

func initClient() (*firestore.Client, *context.Context) {
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
	defer client.Close()
	return client, &ctx
}

// FireFeed is a firestore feed model
type FireFeed struct {
	URL   string `json:"url" firestore:"url"`
	IsRSS bool   `json:"isRSS" firestore:"isRSS"`
}
