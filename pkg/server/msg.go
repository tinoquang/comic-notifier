package server

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/tinoquang/comic-notifier/pkg/conf"
	"github.com/tinoquang/comic-notifier/pkg/model"
	"github.com/tinoquang/comic-notifier/pkg/server/img"
	"github.com/tinoquang/comic-notifier/pkg/store"
	"github.com/tinoquang/comic-notifier/pkg/util"
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

	comic, err := m.subscribeComic(ctx, "psid", senderID, text)
	if err != nil {
		if strings.Contains(err.Error(), "Already") {
			sendTextBack(senderID, "Already subscribed")
		} else {
			sendTextBack(senderID, "Please try again later")
		}
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
			sendTextBack(senderID, "You haven't subscribe this comic yet!")
			return
		}
		return
	}

	sendQuickReplyChoice(senderID, c)
}

// HandleQuickReply handle messages when user click "Yes" to confirm unsubscribe action
func (m *MSG) HandleQuickReply(ctx context.Context, senderID, payload string) {
	comicID, err := strconv.Atoi(payload)

	c, err := m.store.Comic.GetByPSID(ctx, senderID, comicID)
	if err != nil {
		util.Danger(err)
		sendTextBack(senderID, "Please try again later")
		return
	}

	err = m.store.Subscriber.Delete(ctx, senderID, comicID)
	if err != nil {
		sendTextBack(senderID, "Please try again later")
		return
	}

	s, err := m.store.Subscriber.ListByComicID(ctx, comicID)
	if err != nil {
		util.Danger(err)

	}

	if len(s) == 0 {
		img.DeleteImg(string(c.ImgurID))
		m.store.Comic.Delete(ctx, comicID)
	}
	sendTextBack(senderID, fmt.Sprintf("Unsub %s\nSuccess!", c.Name))

}

// Update comic helper
func (m *MSG) subscribeComic(ctx context.Context, field, id, comicURL string) (*model.Comic, error) {

	return m.store.SubscribeComic(ctx, field, id, comicURL)
}
