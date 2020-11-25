package main

import (
	"context"
	"net/http"
	"time"

	firestore "cloud.google.com/go/firestore"
	"github.com/go-chi/chi/middleware"
	"github.com/rs/zerolog/hlog"
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

func accessHandlerFunc(r *http.Request, status, size int, duration time.Duration) {
	hlog.FromRequest(r).Info().
		Str("method", r.Method).
		Str("requestID", middleware.GetReqID(r.Context())).
		Str("source", r.RemoteAddr).
		Stringer("url", r.URL).
		Int("status", status).
		Int("size", size).
		Dur("latency", duration).
		Msg("Incoming Request")
}
