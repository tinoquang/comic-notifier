package server

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/tinoquang/comic-notifier/pkg/conf"
	"github.com/tinoquang/comic-notifier/pkg/store"
)

// MSG -> server handler for messenger endpoint
type MSG struct {
	cfg   *conf.Config
	store *store.Stores
}

// NewMSG return new api interface
func NewMSG(c *conf.Config, s *store.Stores) *MSG {
	return &MSG{cfg: c, store: s}
}

/* Message handler function */

// HandleTxtMsg handle text messages from facebook user
func (m *MSG) HandleTxtMsg(ctx context.Context, senderID, text string) {

	sendActionBack(senderID, "mark_seen")
	sendActionBack(senderID, "typing_on")
	defer sendActionBack(senderID, "typing_off")

	comic, err := subscribeComic(ctx, m.cfg, m.store, "psid", senderID, text)
	if err != nil {
		sendTextBack(senderID, err.Error())
		return
	}

	sendTextBack(senderID, "Subscribed")

	// send back message in template with buttons
	sendNormalReply(senderID, comic)

}

// HandlePostback handle messages when user click "Unsucsribe button"
func (m *MSG) HandlePostback(ctx context.Context, senderID, payload string) {

	sendActionBack(senderID, "mark_seen")
	sendActionBack(senderID, "typing_on")
	defer sendActionBack(senderID, "typing_off")

	comicID, _ := strconv.Atoi(payload)

	c, err := m.store.Comic.GetByPSID(ctx, senderID, comicID)

	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			sendTextBack(senderID, fmt.Sprintf("Comic %s is not subscribed", c.Name))
			return
		}
		return
	}

	sendQuickReplyChoice(senderID, c)
}

// HandleQuickReply handle messages when user click "Yes" to confirm unsubscribe action
func (m *MSG) HandleQuickReply(ctx context.Context, senderID, payload string) {
	comicID, err := strconv.Atoi(payload)

	c, _ := m.store.Comic.GetByPSID(ctx, senderID, comicID)

	err = m.store.Subscriber.Delete(ctx, senderID, comicID)
	if err != nil {
		sendTextBack(senderID, "Please try again later")
	} else {
		sendTextBack(senderID, fmt.Sprintf("Unsubscribe %s", c.Name))
	}
}
