package main

import (
	"github.com/labstack/echo/v4"
	"github.com/tinoquang/comic-notifier/pkg/conf"
	"github.com/tinoquang/comic-notifier/pkg/db"
	"github.com/tinoquang/comic-notifier/pkg/msg"
	"github.com/tinoquang/comic-notifier/pkg/server"
	"github.com/tinoquang/comic-notifier/pkg/store"
)

func main() {

	e := echo.New()

	// Get environment variable
	cfg := conf.New()

	// Connect to DB
	dbconn := db.New(cfg)

	// Init DB handler
	s := store.New(dbconn, cfg)

	// Init main business logic server
	svr := server.New(cfg, s)

	// Facebook webhook
	msg.RegisterHandler(e.Group("/webhook"), cfg, svr)

	// API handler register
	// api.RegisterHandlers(e.Group("/api/v1"), svr)
	// Start the server
	e.Logger.Fatal(e.Start(":8080"))

}
