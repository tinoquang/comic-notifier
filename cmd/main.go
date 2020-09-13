package main

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/tinoquang/comic-notifier/pkg/msg"
)

func main() {

	e := echo.New()

	e.GET("/", hello)

	// Facebook webhook
	msg.RegisterHandler(e.Group("/webhook"))

	// Start the server
	e.Logger.Fatal(e.Start(":8080"))

}

func hello(c echo.Context) error {
	return c.String(http.StatusOK, "Hello world")
}
