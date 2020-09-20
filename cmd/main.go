package main

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/tinoquang/comic-notifier/pkg/conf"
	"github.com/tinoquang/comic-notifier/pkg/db"
	"github.com/tinoquang/comic-notifier/pkg/msg"
)

func main() {

	e := echo.New()

	e.GET("/", hello)

	// Get environment variable
	cfg := conf.New()

	// Connect to DB
	dbconn := db.New(*cfg)

	fmt.Println(dbconn)
	// Facebook webhook
	msg.RegisterHandler(e.Group("/webhook"), cfg)

	// Start the server
	e.Logger.Fatal(e.Start(":8080"))

}

func hello(c echo.Context) error {
	return c.String(http.StatusOK, "Hello world")
}
