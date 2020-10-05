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

	// Page interface
	GetPage(ctx context.Context, name string) (*model.Page, error)

	// Comic interface
	Comics(ctx context.Context) ([]model.Comic, error)
	UpdateComic(ctx context.Context, comic *model.Comic) (bool, error)
	GetComicByUserID(ctx context.Context, userID, comicID int) (*model.Comic, error)
	SubscribeComic(ctx context.Context, field string, id string, comicURL string) (*model.Comic, error)
	UnsubscribeComic(ctx context.Context, userID, comicID int) error

	// User interface
	GetUserByPSID(ctx context.Context, psid string) (*model.User, error)
	GetUserByComicID(ctx context.Context, comicID int) ([]model.User, error)
}

// Handler main handler for incoming HTTP request
type Handler struct {
	svr ServerInterface
}

// RegisterHandler : register webhook handler
func RegisterHandler(g *echo.Group, cfg *conf.Config, svr ServerInterface) {

	// Get env config
	messengerEndpoint = cfg.Webhook.GraphEndpoint + "me/messages"
	webhookToken = cfg.Webhook.WebhookToken
	pageToken = cfg.FBSecret.PakeToken

	// Start worker pool
	go updateThread(svr, cfg.WrkDat.WorkerNum, cfg.WrkDat.Timeout)

	// Create main handler
	h := Handler{svr: svr}

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
		return errors.Wrap(err, "Can't parse message from messenger")
	}

	if m.Object == "page" {
		for _, entry := range m.Entries {
			// Mark message as "seen"
			sendActionBack(entry.Messaging[0].Sender.ID, "mark_seen")

			if len(entry.Messaging) != 0 {
				switch {
				case entry.Messaging[0].PostBack != nil:
					go handlePostback(h.svr, entry.Messaging[0])
				case entry.Messaging[0].Message.QuickReply != nil:
					go handleQuickReply(h.svr, entry.Messaging[0])
				case entry.Messaging[0].Message.Text != "":
					go handleText(h.svr, entry.Messaging[0])
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
