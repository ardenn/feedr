package main

import (
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

func main() {
	// Echo instance
	e := echo.New()
	fire, ctx := initClient()

	// Middleware
	e.Use(FirestoreToContext(fire, ctx))
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "${time_rfc3339_nano} ${level} ${remote_ip} ${method} ${path} ${status} latency: ${latency_human}\n",
	}))
	e.Use(middleware.Recover())

	// Routes
	e.POST("/command", commandHandler)

	// Start server
	e.Logger.Fatal(e.Start(":8000"))
}
