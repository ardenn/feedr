package main

import (
	"context"

	firestore "cloud.google.com/go/firestore"
	"github.com/labstack/echo"
)

// CustomContext model
type CustomContext struct {
	echo.Context
	fire *firestore.Client
	ctx  *context.Context
}

// FirestoreToContext middleware
func FirestoreToContext(fire *firestore.Client, ctx *context.Context) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			cc := &CustomContext{c, fire, ctx}
			return next(cc)
		}

	}
}
