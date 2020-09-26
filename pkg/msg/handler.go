package msg

import (
	"context"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	"github.com/tinoquang/comic-notifier/pkg/conf"
	"github.com/tinoquang/comic-notifier/pkg/model"
	"github.com/tinoquang/comic-notifier/pkg/util"
)

var (
	messengerEndpoint string
	pageToken         string
	webhookToken      string
)

// ServerInterface contain all server's method
type ServerInterface interface {
	GetPage(ctx context.Context, name string) (*model.Page, error)
	SubscribeComic(ctx context.Context, field string, id string, comicURL string) (*model.Comic, error)
}

// RequestHandler main handler for incoming HTTP request
type RequestHandler struct {
	svr ServerInterface
}

type msgHandler struct {
	svr ServerInterface
	req Messaging
}

func newMsgHandler(svr ServerInterface, req Messaging) *msgHandler {
	return &msgHandler{
		svr: svr,
		req: req,
	}
}

// RegisterHandler : register webhook handler
func RegisterHandler(g *echo.Group, cfg *conf.Config, svr ServerInterface) {

	messengerEndpoint = cfg.Webhook.GraphEndpoint + "me/messages"
	webhookToken = cfg.Webhook.WebhookToken
	pageToken = cfg.FBSecret.PakeToken

	h := RequestHandler{svr: svr}

	// Webhook verify message
	g.GET("", h.verifyWebhook)

	// Handle user message
	g.POST("", h.parseUserMsg)

}

func (h *RequestHandler) verifyWebhook(c echo.Context) error {

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

func (h *RequestHandler) parseUserMsg(c echo.Context) error {

	m := &UserMessage{}

	if err := c.Bind(m); err != nil {
		return errors.Wrap(err, "Can't parse message from messenger")
	}

	if m.Object == "page" {
		for _, entry := range m.Entries {
			mh := newMsgHandler(h.svr, entry.Messaging[0])

			// Mark message as "seen"
			mh.sendActionBack("mark_seen")

			if len(entry.Messaging) != 0 {
				switch {
				case entry.Messaging[0].PostBack != nil:
					util.Info("postback")
				// 	go handlePostBack(&entry.Messaging[0])
				case entry.Messaging[0].Message.QuickReply != nil:
					util.Info("quick reply")
				// 		go returnToQuickReply(&entry.Messaging[0])
				case entry.Messaging[0].Message.Text != "":
					go mh.handleText()
				default:
					util.Warning("Only support text, postback and quick-reply !!!")
				}
			} else {
				util.Warning("Messge from messenger is empty !")
			}
		}
	} else {
		util.Warning("Message request unknown!!!")
	}

	return c.String(http.StatusOK, "")
}
