package main

import (
	"os"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

func main() {
	// Echo instance
	e := echo.New()
	fire := initClient()
	defer fire.Close()

	// Middleware
	e.Use(FirestoreToContext(fire))
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "${time_rfc3339_nano} ${level} ${remote_ip} ${method} ${path} ${status} latency: ${latency_human}\n",
	}))
	e.Use(middleware.Recover())

	// Routes
	e.POST("/command", commandHandler)
	e.GET("/crawl", crawlHandler)

	// Start server
	e.Logger.Fatal(e.Start(":" + os.Getenv("PORT")))
}
