package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/rs/zerolog/hlog"
	"github.com/rs/zerolog/log"
)

func main() {
	// Mux instance
	r := chi.NewRouter()
	fire := initClient()
	defer fire.Close()

	// Middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	// r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(hlog.NewHandler(log.Logger))
	r.Use(hlog.AccessHandler(func(r *http.Request, status, size int, duration time.Duration) {
		hlog.FromRequest(r).Info().
			Str("method", r.Method).
			Str("requestID", middleware.GetReqID(r.Context())).
			Str("source", r.RemoteAddr).
			Stringer("url", r.URL).
			Int("status", status).
			Int("size", size).
			Dur("latency", duration).
			Msg("Incoming Request")
	}))
	r.Use(FirestoreToContext(fire))

	// Routes
	r.Post("/command", commandHandler)
	r.Get("/crawl", crawlHandler)

	// Start server
	fmt.Println(banner)
	fmt.Println("Version: ", version)
	fmt.Println("Server started on port " + os.Getenv("PORT") + " ...")
	if err := http.ListenAndServe(":"+os.Getenv("PORT"), r); err != nil {
		log.Fatal().Err(err).Msg("Startup failed")
	}
}
