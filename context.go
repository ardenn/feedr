package main

import (
	firestore "cloud.google.com/go/firestore"
	"github.com/labstack/echo"
)

// CustomContext model
type CustomContext struct {
	echo.Context
	fire *firestore.Client
}

// FirestoreToContext middleware
func FirestoreToContext(fire *firestore.Client) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			cc := &CustomContext{c, fire}
			return next(cc)
		}

	}
}
