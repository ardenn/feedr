package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-pg/pg/v10"
	"github.com/rs/zerolog/hlog"
	"github.com/rs/zerolog/log"
)

var db *pg.DB

func main() {
	// Mux instance
	r := chi.NewRouter()
	db = dbConnect()
	logQueries(db)
	defer db.Close()

	// Middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(hlog.NewHandler(log.Logger))
	r.Use(hlog.AccessHandler(accessHandlerFunc))

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
