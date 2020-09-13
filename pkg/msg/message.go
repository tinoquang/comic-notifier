package msg

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func RegisterHandler(g *echo.Group) {

	// Webhook verify message
	g.GET("", verifyWebhook)

}

func verifyWebhook(c echo.Context) error {

	params := c.QueryString()

	if params == "" {
		return c.String(http.StatusBadRequest, "No query params")
	}

	mode := c.QueryParam("hub.mode")
	token := c.QueryParam("hub.verify_token")
	challenge := c.QueryParam("hub.challenge")

	if mode == "subscribe" && token == "quangmt2" {
		return c.String(http.StatusOK, challenge)
	}

	return c.String(http.StatusBadRequest, "Invalid token")
}
