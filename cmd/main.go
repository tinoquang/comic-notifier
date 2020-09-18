package main

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/tinoquang/comic-notifier/pkg/conf"
	"github.com/tinoquang/comic-notifier/pkg/msg"
)

func main() {

	e := echo.New()

	e.GET("/", hello)

	// Get environment variable
	cfg := conf.New()

	// Facebook webhook
	msg.RegisterHandler(e.Group("/webhook"), cfg)

	fmt.Println(cfg.Webhook.WebhookToken)
	// Start the server
	e.Logger.Fatal(e.Start(":8080"))

}

func hello(c echo.Context) error {
	return c.String(http.StatusOK, "Hello world")
}
