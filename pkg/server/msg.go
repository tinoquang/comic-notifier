package server

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/tinoquang/comic-notifier/pkg/logging"
	"github.com/tinoquang/comic-notifier/pkg/model"
	"github.com/tinoquang/comic-notifier/pkg/server/img"
	"github.com/tinoquang/comic-notifier/pkg/store"
	"github.com/tinoquang/comic-notifier/pkg/util"
)

// MSG -> server handler for messenger endpoint
type MSG struct {
	store *store.Stores
}

// NewMSG return new api interface
func NewMSG(s *store.Stores) *MSG {
	return &MSG{store: s}
}

/* Message handler function */

// HandleTxtMsg handle text messages from facebook user
func (m *MSG) HandleTxtMsg(ctx context.Context, senderID, text string) {

	sendActionBack(senderID, "mark_seen")
	sendActionBack(senderID, "typing_on")
	defer sendActionBack(senderID, "typing_off")

	if text[0] == '/' {
		responseCommand(ctx, senderID, text)
		return
	}

	comic, err := m.subscribeComic(ctx, senderID, text)
	if err != nil {
		if err == util.ErrAlreadySubscribed {
			sendTextBack(senderID, "Already subscribed")
		} else if strings.Contains(err.Error(), "too fast") {
			// Upload image API is busy
			sendTextBack(senderID, "Hiện tại tôi đang busy, hãy thử lại sau nhé! :)") // handle later: get time delay and send back to user
		} else if err == util.ErrPageNotSupported {
			sendTextBack(senderID, "Trang truyện hiện tại chưa hỗ trợ !!!")
			responseCommand(ctx, senderID, "/page")
		} else if err == util.ErrInvalidURL {
			sendTextBack(senderID, "Đường dẫn chưa chính xác, hãy xem qua hướng dẫn bằng lệnh /tutor")
			// responseCommand(ctx, senderID, "/help")
		} else {
			sendTextBack(senderID, "Hiện tại tôi đang busy, hãy thử lại sau nhé! :)")
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

	if strings.Contains(payload, "get-started") {
		reponseGetStarted(ctx, senderID, payload)
		return
	}

	comicID, _ := strconv.Atoi(payload)

	c, err := m.store.Comic.CheckComicSubscribe(ctx, senderID, comicID)

	if err != nil {
		if err == util.ErrNotFound {
			sendTextBack(senderID, "This comic is not subscribed yet!")
			return
		}
		return
	}

	sendQuickReplyChoice(senderID, c)
}

// HandleQuickReply handle messages when user click "Yes" to confirm unsubscribe action
func (m *MSG) HandleQuickReply(ctx context.Context, senderID, payload string) {
	comicID, err := strconv.Atoi(payload)

	c, err := m.store.Comic.CheckComicSubscribe(ctx, senderID, comicID)
	if err != nil {
		logging.Danger(err)
		sendTextBack(senderID, "This comic is not subscribed yet!")
		return
	}

	err = m.store.Subscriber.Delete(ctx, senderID, comicID)
	if err != nil {
		sendTextBack(senderID, "Please try again later")
		return
	}

	s, err := m.store.Subscriber.ListByComicID(ctx, comicID)
	if err != nil {
		logging.Danger(err)

	}

	if len(s) == 0 {
		img.DeleteFirebaseImg(c.Page, c.Name)
		m.store.Comic.Delete(ctx, comicID)
	}
	sendTextBack(senderID, fmt.Sprintf("Unsub %s\nSuccess!", c.Name))

}

// Update comic helper
func (m *MSG) subscribeComic(ctx context.Context, id, comicURL string) (*model.Comic, error) {

	return m.store.SubscribeComic(ctx, id, comicURL)
}

func responseCommand(ctx context.Context, senderID, text string) {

	if text == "/list" {
		sendTextBack(senderID, "Xem danh sách truyện đã đăng kí ở đường dẫn sau:")
		sendTextBack(senderID, "https://comicnotifier.herokuapp.com")
	} else if text == "/page" {
		sendTextBack(senderID, "Hiện tôi hỗ trợ các trang: beeng.net, blogtruyen.vn, truyenhtranh.net và truyentranhtuan.com")
	} else if text == "/tutor" {
		sendTextBack(senderID, "Xem hướng dẫn tại đây:")
		sendTextBack(senderID, "https://comicnotifier.herokuapp.com/tutorial")
	} else {
		sendTextBack(senderID, `Các lệnh tối hỗ trợ:
- /list:  xem các truyện đã đăng kí
- /page:  xem các trang web hiện tại BOT hỗ trợ
- /tutor: xem hướng dẫn
- /help:  xem lại các lệnh hỗ trợ`)
	}
	return
}

func reponseGetStarted(ctx context.Context, senderID, payload string) {

	sendTextBack(senderID, "Welcome to Cominify!")
	sendTextBack(senderID, "Tôi là chatbot giúp theo dõi truyện tranh và thông báo mỗi khi truyện có chapter mới")
	sendTextBack(senderID, `Các lệnh tối hỗ trợ:
- /list:  xem các truyện đã đăng kí
- /page:  xem các trang web hiện tại BOT hỗ trợ
- /tutor: xem hướng dẫn
- /help:  xem lại các lệnh hỗ trợ`)
	return
}
