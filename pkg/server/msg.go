package server

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/asaskevich/govalidator"
	db "github.com/tinoquang/comic-notifier/pkg/db/sqlc"
	"github.com/tinoquang/comic-notifier/pkg/logging"
	"github.com/tinoquang/comic-notifier/pkg/util"
)

// MSG -> server handler for messenger endpoint
type MSG struct {
	sync.Mutex
	store   db.Store
	crawler infoCrawler
}

// NewMSG return new api interface
func NewMSG(s db.Store, crwl infoCrawler) *MSG {
	return &MSG{store: s, crawler: crwl}
}

/* Message handler function */

// HandleTxtMsg handle text messages from facebook user
func (m *MSG) HandleTxtMsg(ctx context.Context, senderID, text string) {

	sendActionBack(senderID, "mark_seen")
	sendActionBack(senderID, "typing_on")
	defer sendActionBack(senderID, "typing_off")

	if text[0] == '/' {
		m.responseCommand(ctx, senderID, text)
		return
	}

	if valid := govalidator.IsURL(text); !valid {
		sendTextBack(senderID, "Cú pháp chưa chính xác")
		m.responseCommand(ctx, senderID, "")
		return
	}

	comic, err := m.SubscribeComic(ctx, senderID, text)
	if err != nil {
		if err == util.ErrAlreadySubscribed {
			sendTextBack(senderID, fmt.Sprintf("%s đã được đăng ký, BOT sẽ thông báo cho bạn khi có chương mới", comic.Name))
		} else if strings.Contains(err.Error(), "too fast") || err == util.ErrCrawlTimeout {
			// Upload image API is busy
			sendTextBack(senderID, "Đăng ký không thành công, hãy thử lại sau nhé!") // handle later: get time delay and send back to user
		} else if err == util.ErrPageNotSupported {
			sendTextBack(senderID, "Trang truyện này chưa được hỗ trợ, dùng lệnh /page để xem các trang tôi hỗ trợ")
			m.responseCommand(ctx, senderID, "/page")
		} else if err == util.ErrInvalidURL {
			sendTextBack(senderID, "Đường dẫn chưa chính xác, hãy xem qua hướng dẫn bằng lệnh /tutor")
		} else {
			sendTextBack(senderID, "Đăng ký không thành công, hãy thử lại sau nhé")
		}
		return
	}

	// send back message in template with bDnDwauttons
	sendTextBack(senderID, fmt.Sprintf("Đăng ký truyện %s thành công", comic.Name))
	sendActionBack(senderID, "typing_on")
	delayMS(500)
	sendNormalReply(senderID, comic)

}

// HandlePostback handle messages when user click "Unsubsribe button"
func (m *MSG) HandlePostback(ctx context.Context, senderID, payload string) {

	sendActionBack(senderID, "mark_seen")
	sendActionBack(senderID, "typing_on")
	defer sendActionBack(senderID, "typing_off")

	if strings.Contains(payload, "get-started") {
		m.reponseGetStarted(ctx, senderID)
		return
	}

	if payload[0] == '/' {
		m.responseCommand(ctx, senderID, payload)
		return
	}

	comicID, _ := strconv.Atoi(payload)
	comic, err := m.store.GetComicByPSIDAndComicID(ctx, db.GetComicByPSIDAndComicIDParams{
		Psid: sql.NullString{String: senderID, Valid: true},
		ID:   int32(comicID),
	})
	if err != nil {
		if err == sql.ErrNoRows {
			sendTextBack(senderID, "Truyện chưa được đăng ký")
			return
		}

		logging.Danger(err)
		sendTextBack(senderID, "Hiện tại server đang busy, bạn hãy đợi một lát rồi thử lại nhé")
		return
	}

	sendQuickReplyChoice(senderID, comic)
}

// HandleQuickReply handle messages when user click "Yes" to confirm unsubscribe action
func (m *MSG) HandleQuickReply(ctx context.Context, senderID, payload string) {

	sendActionBack(senderID, "mark_seen")
	sendActionBack(senderID, "typing_on")
	defer sendActionBack(senderID, "typing_off")

	if payload == "Not unsub" {
		sendActionBack(senderID, "mark_seen")
		return
	}

	if payload[0] == '/' {
		m.responseCommand(ctx, senderID, payload)
		return
	}

	comicID, err := strconv.Atoi(payload)
	if err != nil {
		logging.Danger(err)
		return
	}

	user, err := m.store.GetUserByPSID(ctx, sql.NullString{String: senderID, Valid: true})
	if err != nil {
		logging.Danger(err)
		sendTextBack(senderID, "Truyện chưa được đăng ký")
		return
	}

	c, err := m.store.GetComicByPSIDAndComicID(ctx, db.GetComicByPSIDAndComicIDParams{
		Psid: sql.NullString{String: senderID, Valid: true},
		ID:   int32(comicID),
	})
	if err != nil {
		logging.Danger(err)
		sendTextBack(senderID, "Truyện chưa được đăng ký")
		return
	}

	err = m.store.DeleteSubscriber(ctx, db.DeleteSubscriberParams{
		UserID:  user.ID,
		ComicID: int32(comicID),
	})

	if err != nil {
		sendTextBack(senderID, "Hiện tại server đang busy, bạn hãy đợi một lát rồi thử lại nhé")
		return
	}

	// Check if any user still subscribed to this comic, if not remove comic from DB
	users, err := m.store.ListUsersPerComic(ctx, c.ID)
	if err != nil {
		logging.Danger(err)
		sendTextBack(senderID, "Hiện tại server đang busy, bạn hãy đợi một lát rồi thử lại nhé")
		return
	}

	if len(users) == 0 {
		m.store.RemoveComic(ctx, c.ID)
	}
	sendTextBack(senderID, fmt.Sprintf("Hủy đăng ký %s thành công", c.Name))

}

func (m *MSG) responseCommand(ctx context.Context, senderID, text string) {

	switch text {
	case "/list":
		user, err := m.store.GetUserByPSID(ctx, sql.NullString{String: senderID, Valid: true})
		if err != nil {
			sendTutor(senderID)
			return
		}

		comics, err := m.store.ListComicsPerUser(ctx, user.ID)
		if err != nil || len(comics) == 0 {
			sendTutor(senderID)
		} else {
			sendTextBack(senderID, fmt.Sprintf("Bạn đã đăng ký nhận thông báo cho %d truyện", len(comics)))
			sendTextBack(senderID, `Xem chi tiết tại
www.cominify-bot.xyz`)
		}
	case "/page":
		sendTextBack(senderID, `Các trạng hiện tại tôi hỗ trợ:
beeng.net
blogtruyen.vn
truyentranhtuan.com
truyenqq.com
hocvientruyentranh.net`)
	case "/tutor":
		sendTextBack(senderID, "Để đăng kí, chỉ cần gởi cho BOT link truyện bạn muốn nhận thông báo")
		sendTextBack(senderID, "Ví dụ bạn muốn đăng ký truyện One Piece ở trang blogtruyen.vn, hãy gởi cho BOT đường link sau:")
		sendTextBack(senderID, "https://blogtruyen.vn/139/one-piece")
		sendTextBack(senderID, `Hãy thử copy đường link trên và gởi cho BOT, nếu vẫn chưa rõ bạn có thể xem hướng dẫn tại
www.cominify-bot.xyz/tutorial`)
	default:
		sendSupportCommand(senderID)
	}

}

func (m *MSG) reponseGetStarted(ctx context.Context, senderID string) {

	sendTextBack(senderID, "Welcome to Comic Notify Bot!")
	sendTextBack(senderID, "Tôi là chatbot giúp theo dõi truyện tranh và thông báo mỗi khi truyện có chapter mới")
	sendSupportCommand(senderID)
}

// SubscribeComic add comic and user to DB
func (m *MSG) SubscribeComic(ctx context.Context, userPSID, comicURL string) (*db.Comic, error) {

	var (
		err   error
		comic db.Comic
		user  db.User
	)

	m.Lock()
	defer m.Unlock()
	comic, err = m.store.GetComicByURL(ctx, comicURL)
	if err != nil {

		if err != sql.ErrNoRows {
			logging.Danger(err)
			return nil, err
		}
		// Comic is not in DB, need to get it's info using crawler pkg
		comic, err = m.crawler.GetComicInfo(ctx, comicURL, false)
		if err != nil {
			logging.Danger(err)
			return nil, err
		}
	}

	// Verify comic again to avoid multiple URL represents same comic, by checking Page + Comic Name
	c, err := m.store.GetComicByPageAndComicName(ctx, db.GetComicByPageAndComicNameParams{
		Page: comic.Page,
		Name: comic.Name,
	})

	if err == nil {
		comic.ID = c.ID
	}

	user, err = m.store.GetUserByPSID(ctx, sql.NullString{String: userPSID, Valid: true})
	if err != nil {

		if err != sql.ErrNoRows {
			logging.Danger(err)
			return nil, err
		}

		user, err = m.crawler.GetUserInfoFromFacebook("psid", userPSID)
		if err != nil {
			logging.Danger(err)
			return nil, err
		}
	}

	_, err = m.store.GetSubscriber(ctx, db.GetSubscriberParams{
		UserID:  user.ID,
		ComicID: comic.ID,
	})

	if err == nil {
		return &comic, util.ErrAlreadySubscribed
	}

	if err != sql.ErrNoRows {
		logging.Danger(err)
		return nil, err
	}

	err = m.store.SubscribeComic(ctx, &comic, &user)
	return &comic, err
}
