package main

import (
	"github.com/labstack/echo/v4"
)

func main() {

	e := echo.New()

	// Facebook webhook
	FBHook := e.Group("/webook")

	msg.RegisterHandler(FBhook)
}
