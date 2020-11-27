package main

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/middleware"
	"github.com/rs/zerolog/hlog"
)

// MiddlewareFunc defines the format of a middleware function
type MiddlewareFunc func(next http.Handler) http.Handler

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
