package msg

import (
	"context"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/tinoquang/comic-notifier/pkg/conf"
	"github.com/tinoquang/comic-notifier/pkg/logging"
)

var webhookToken string

// ServerInterface contain all server's method
type ServerInterface interface {

	// Handler msg interface
	HandleTxtMsg(ctx context.Context, senderID string, text string)
	HandlePostback(ctx context.Context, senderID, payload string)
	HandleQuickReply(ctx context.Context, senderID, payload string)
}

// Handler main handler for incoming HTTP request
type Handler struct {
	cfg *conf.Config
	svi ServerInterface
}

// RegisterHandler : register webhook handler
func RegisterHandler(g *echo.Group, cfg *conf.Config, svi ServerInterface) {

	webhookToken = cfg.Webhook.WebhookToken

	// Create main handler
	h := Handler{cfg: cfg, svi: svi}

	// Register endpoint to handler
	// Webhook verify message
	g.GET("", h.verifyWebhook)

	// Handle user message
	g.POST("", h.parseUserMsg)

}

func (h *Handler) verifyWebhook(c echo.Context) error {

	params := c.QueryString()

	if params == "" {
		return c.String(http.StatusBadRequest, "No query params")
	}

	mode := c.QueryParam("hub.mode")
	token := c.QueryParam("hub.verify_token")
	challenge := c.QueryParam("hub.challenge")

	if mode == "subscribe" && token == webhookToken {
		return c.String(http.StatusOK, challenge)
	}

	return c.String(http.StatusBadRequest, "Invalid token")
}

func (h *Handler) parseUserMsg(c echo.Context) error {

	m := &UserMessage{}

	// Parsing request
	if err := c.Bind(m); err != nil {
		return c.NoContent(http.StatusBadRequest)
	}

	if m.Object == "page" {
		for _, entry := range m.Entries {
			if len(entry.Messaging) != 0 {
				switch {
				case entry.Messaging[0].PostBack != nil:
					go h.handlePostback(entry.Messaging[0], h.cfg.CtxTimeout)
				case entry.Messaging[0].Message.QuickReply != nil:
					go h.handleQuickReply(entry.Messaging[0], h.cfg.CtxTimeout)
				case entry.Messaging[0].Message.Text != "":
					go h.handleText(entry.Messaging[0], h.cfg.CtxTimeout)
				default:
					logging.Warning("Only support text, postback and quick-reply !!!")
				}
			} else {
				logging.Warning("Messge from messenger is empty !")
			}
		}
	} else {
		logging.Warning("Message request unknown!!!")
	}

	return c.NoContent(http.StatusOK)
}
