package main

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/tinoquang/comic-notifier/pkg/api"
	"github.com/tinoquang/comic-notifier/pkg/auth"
	"github.com/tinoquang/comic-notifier/pkg/conf"
	"github.com/tinoquang/comic-notifier/pkg/db"
	"github.com/tinoquang/comic-notifier/pkg/msg"
	"github.com/tinoquang/comic-notifier/pkg/server"
	"github.com/tinoquang/comic-notifier/pkg/store"
)

func main() {

	e := echo.New()

	// Get environment variable
	cfg := conf.New("")

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
	apiGroup.Use(middleware.JWTWithConfig(middleware.JWTConfig{
		SigningKey:  []byte(cfg.JWT.SecretKey),
		Claims:      &jwt.StandardClaims{},
		TokenLookup: "cookie:_session",
	}))
	api.RegisterHandlers(apiGroup, svr.API)

	/* Routing */
	e.Static("/static", "ui/static")
	e.Static("/assets", "ui/assets")

	e.GET("/", func(c echo.Context) error {
		return c.File("ui/index.html")
	})

	// Authentication JWT
	auth.RegisterHandler(e.Group(""), cfg, s)

	// Start the server
	e.Logger.Fatal(e.Start(":" + cfg.Port))

}
