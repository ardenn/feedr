package main

// import (
// 	"cloud.google.com/go/firestore"
// 	"github.com/labstack/echo"
// 	"google.golang.org/api/iterator"
// )

// func fetchFeeds(fire *firestore.Client, c echo.Context) {
// 	var users []FireUser
// 	iter := fire.Collection("userFeeds").Documents(c.Request().Context())
// 	for {
// 		doc, err := iter.Next()
// 		if err == iterator.Done {
// 			break
// 		}
// 		if err != nil {
// 			c.Logger().Errorf("Error reading crawlFeeds", err)
// 		}
// 		user := FireUser{}
// 		if err := doc.DataTo(&user); err != nil {
// 			c.Logger().Errorf("Error reading firestore users list", err)
// 		}
// 	}
// }
