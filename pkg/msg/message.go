package message

import (
	"github.com/labstack/echo/v4"
)

func RegisterHandler(g echo.Group) {

	// Webhook verify message
	g.GET("/", verifyWebhook)

}

func verifyWebhook(ctx echo.Context) (err error) {

	return
}
