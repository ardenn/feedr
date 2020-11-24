package main

import (
	"context"
	"net/http"

	firestore "cloud.google.com/go/firestore"
)

// FireContextKey is the firebase client context.Context key
type FireContextKey string

var fireCtxKey FireContextKey = "firestore"

// MiddlewareFunc defines the format of a middleware function
type MiddlewareFunc func(next http.Handler) http.Handler

// FirestoreToContext middleware
func FirestoreToContext(fire *firestore.Client) MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), fireCtxKey, fire)
			r = r.WithContext(ctx)
			// Call the next handler, which can be another middleware in the chain, or the final handler.
			next.ServeHTTP(w, r)
		})
	}
}

// FireContext finds the fire client from the context. REQUIRES Middleware to have run.
func FireContext(ctx context.Context) *firestore.Client {
	raw, _ := ctx.Value(fireCtxKey).(*firestore.Client)
	return raw
}
