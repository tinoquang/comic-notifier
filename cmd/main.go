package main

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/tinoquang/comic-notifier/pkg/api"
	"github.com/tinoquang/comic-notifier/pkg/auth"
	"github.com/tinoquang/comic-notifier/pkg/conf"
	"github.com/tinoquang/comic-notifier/pkg/crawler"
	db "github.com/tinoquang/comic-notifier/pkg/db/sqlc"
	"github.com/tinoquang/comic-notifier/pkg/msg"
	"github.com/tinoquang/comic-notifier/pkg/server"
)

func main() {

	e := echo.New()
	e.Pre(middleware.RemoveTrailingSlash())

	// Init global config
	conf.Init()

	dbconn := db.NewDBConn()

	crawler := crawler.NewCrawler()

	// Init Repository
	store := db.NewStore(dbconn, crawler)

	// Init main business logic server
	svr := server.New(store)

	// // Facebook webhook
	msg.RegisterHandler(e.Group("/webhook"), svr.Msg)

	// API handler register
	apiGroup := e.Group("/api/v1")
	apiGroup.Use(middleware.JWTWithConfig(middleware.JWTConfig{
		SigningKey:  []byte(conf.Cfg.JWT.SecretKey),
		Claims:      &jwt.StandardClaims{},
		TokenLookup: "cookie:_session",
	}))
	api.RegisterHandlers(apiGroup, svr.API)

	/* Routing */
	e.Static("/static", "ui/static")
	e.Static("/assets", "ui/assets")
	e.Static("/favicon.ico", "ui/favicon.ico")

	e.GET("/*", func(c echo.Context) error {
		return c.File("ui/index.html")
	})

	// Authentication JWT
	auth.RegisterHandler(e.Group(""), store)

	// Start the server
	e.Logger.Fatal(e.Start(":" + conf.Cfg.Port))

}
