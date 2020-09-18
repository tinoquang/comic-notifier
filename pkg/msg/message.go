package msg

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/tinoquang/comic-notifier/pkg/conf"
	"github.com/tinoquang/comic-notifier/pkg/util"
)

var (
	messengerEndpoint string
	pageToken         string
	webhookToken      string
)

// RegisterHandler : register webhook handler
func RegisterHandler(g *echo.Group, cfg *conf.Config) {

	messengerEndpoint = cfg.Webhook.MessengerEndpoint
	webhookToken = cfg.Webhook.WebhookToken
	pageToken = cfg.FBSecret.PakeToken

	// Webhook verify message
	g.GET("", verifyWebhook)

	// Handle user message
	g.POST("", userMsgHandler)

}

func verifyWebhook(c echo.Context) error {

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

func userMsgHandler(c echo.Context) error {

	m := &UserMessage{}

	if err := c.Bind(m); err != nil {
		util.Danger(err)
		return err
	}

	if m.Object == "page" {
		for _, entry := range m.Entries {
			msg := entry.Messaging[0]
			sendActionBack(msg.Sender.ID, "mark_seen")

			if len(entry.Messaging) != 0 {
				switch {
				case entry.Messaging[0].PostBack != nil:
					util.Info("postback")
				// 	go handlePostBack(&entry.Messaging[0])
				case entry.Messaging[0].Message.QuickReply != nil:
					util.Info("quick reply")
				// 		go returnToQuickReply(&entry.Messaging[0])
				case entry.Messaging[0].Message.Text != "":
					go msg.textHandler()
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
