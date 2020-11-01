package main

import (
	"github.com/labstack/echo/v4"
	"github.com/tinoquang/comic-notifier/pkg/api"
	"github.com/tinoquang/comic-notifier/pkg/auth"
	"github.com/tinoquang/comic-notifier/pkg/conf"
	"github.com/tinoquang/comic-notifier/pkg/db"
	"github.com/tinoquang/comic-notifier/pkg/mdw"
	"github.com/tinoquang/comic-notifier/pkg/msg"
	"github.com/tinoquang/comic-notifier/pkg/server"
	"github.com/tinoquang/comic-notifier/pkg/store"
)

func main() {

	e := echo.New()

	// Get environment variable
	cfg := conf.New("")

	// Set middleware config
	mdw.SetConfig(cfg)

	// Connect to DB
	dbconn := db.New(cfg)

	// Init DB handler
	s := store.New(dbconn, cfg)

	// Init main business logic server
	svr := server.New(cfg, s)

	// Facebook webhook
	msg.RegisterHandler(e.Group("/webhook"), cfg, svr.Msg)

	// API handler register
	apiGroup := e.Group("/api/v1")
	apiGroup.Use(mdw.CheckLoginStatus)
	api.RegisterHandlers(apiGroup, svr.API)

	/* Routing */
	e.Static("/static", "ui/static")
	e.GET("/", func(c echo.Context) error {
		return c.File("ui/index.html")
	}, mdw.CheckLoginStatus)

	// Authentication JWT
	auth.RegisterHandler(e.Group(""), cfg)

	// Start the server
	e.Logger.Fatal(e.Start(":" + cfg.Port))

}
