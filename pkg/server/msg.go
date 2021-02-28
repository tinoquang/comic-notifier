package server

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"

	db "github.com/tinoquang/comic-notifier/pkg/db/sqlc"
	"github.com/tinoquang/comic-notifier/pkg/logging"
	"github.com/tinoquang/comic-notifier/pkg/util"
)

// MSG -> server handler for messenger endpoint
type MSG struct {
	store db.Stores
}

// NewMSG return new api interface
func NewMSG(s db.Stores) *MSG {
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

	comic, err := m.store.SubscribeComic(ctx, senderID, text)
	if err != nil {
		if err == util.ErrAlreadySubscribed {
			sendTextBack(senderID, fmt.Sprintf("%s đã được đăng ký", comic.Name))
		} else if strings.Contains(err.Error(), "too fast") {
			// Upload image API is busy
			sendTextBack(senderID, "Đăng ký không thành công, hãy thử lại sau nhé!") // handle later: get time delay and send back to user
		} else if err == util.ErrPageNotSupported {
			sendTextBack(senderID, "Trang truyện hiện tại chưa hỗ trợ !!!")
			responseCommand(ctx, senderID, "/page")
		} else if err == util.ErrInvalidURL {
			sendTextBack(senderID, "Đường dẫn chưa chính xác, hãy xem qua hướng dẫn bằng lệnh /tutor")
			// responseCommand(ctx, senderID, "/help")
		} else {
			sendTextBack(senderID, "Đăng ký không thành công, hãy thử lại sau nhé!")
		}
		return
	}

	sendTextBack(senderID, fmt.Sprintf("Đăng ký truyện %s", comic.Name))

	// send back message in template with bDnDwauttons
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
	comic, err := m.store.GetComicByPSIDAndComicID(ctx, db.GetComicByPSIDAndComicIDParams{
		Psid: sql.NullString{String: senderID, Valid: true},
		ID:   int32(comicID),
	})
	if err != nil {
		if err == sql.ErrNoRows {
			sendTextBack(senderID, "Truyện chưa được đăng ký !")
			return
		}

		logging.Danger(err)
		sendTextBack(senderID, "Đợi xíu rồi thử lại nhé bạn")
		return
	}

	sendQuickReplyChoice(senderID, comic)
}

// HandleQuickReply handle messages when user click "Yes" to confirm unsubscribe action
func (m *MSG) HandleQuickReply(ctx context.Context, senderID, payload string) {
	comicID, err := strconv.Atoi(payload)

	user, err := m.store.GetUserByPSID(ctx, sql.NullString{String: senderID, Valid: true})
	if err != nil {
		logging.Danger(err)
		sendTextBack(senderID, "Truyện chưa được đăng ký !")
		return
	}

	c, err := m.store.GetComicByPSIDAndComicID(ctx, db.GetComicByPSIDAndComicIDParams{
		Psid: sql.NullString{String: senderID, Valid: true},
		ID:   int32(comicID),
	})
	if err != nil {
		logging.Danger(err)
		sendTextBack(senderID, "Truyện chưa được đăng ký !")
		return
	}

	err = m.store.DeleteSubscriber(ctx, db.DeleteSubscriberParams{
		UserID:  user.ID,
		ComicID: int32(comicID),
	})

	if err != nil {
		sendTextBack(senderID, "Please try again later")
		return
	}

	// Check if any user still subscribed to this comic, if not remove comic from DB
	users, err := m.store.ListUsersPerComic(ctx, c.ID)
	if err != nil {
		logging.Danger(err)
		sendTextBack(senderID, "Đợi xíu rồi thử lại nhé bạn")
		return
	}

	if len(users) == 0 {
		m.store.DeleteComic(ctx, c.ID)
	}
	sendTextBack(senderID, fmt.Sprintf("Đã hủy đăng ký %s!", c.Name))

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
