package server

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

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
func delayMS(second int) {
	time.Sleep(time.Duration(second) * time.Millisecond)
}

// HandleTxtMsg handle text messages from facebook user
func (m *MSG) HandleTxtMsg(ctx context.Context, senderID, text string) {

	sendActionBack(senderID, "mark_seen")
	sendActionBack(senderID, "typing_on")
	defer sendActionBack(senderID, "typing_off")

	if text[0] == '/' {
		m.responseCommand(ctx, senderID, text)
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
			sendTextBack(senderID, "Trang truyện hiện tại chưa hỗ trợ")
			m.responseCommand(ctx, senderID, "/page")
		} else if err == util.ErrInvalidURL {
			sendTextBack(senderID, "Đường dẫn chưa chính xác, hãy xem qua hướng dẫn bằng lệnh /tutor")
		} else {
			sendTextBack(senderID, "Đăng ký không thành công, hãy thử lại sau nhé")
		}
		return
	}

	// send back message in template with bDnDwauttons
	sendNormalReply(senderID, comic)
	delayMS(500)
	sendTextBack(senderID, fmt.Sprintf("Đăng ký truyện %s thành công", comic.Name))
	delayMS(500)
	sendTextBack(senderID, "Nếu muốn hủy nhận thông báo cho truyện này, click Hủy đăng ký ở trên")
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

	if payload == "Not unsub" {
		sendActionBack(senderID, "mark_seen")
		return
	}

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
	sendTextBack(senderID, fmt.Sprintf("Đã hủy đăng ký truyện %s", c.Name))

}

func (m *MSG) responseCommand(ctx context.Context, senderID, text string) {

	switch text {
	case "/start":
		sendTextBack(senderID, "Hướng dân đăng kí truyện")
		sendTextBack(senderID, "Ví dụ : Bạn muốn nhận thông báo cho truyện Onepiece ở trạng blogtruyen.vn")
		sendTextBack(senderID, "Copy đường dẫn sau và gởi cho BOT")
		sendTextBack(senderID, "blogtruyen.vn/139/one-piece")
	case "/list":
		userID, err := strconv.Atoi(senderID)
		if err != nil {
			logging.Danger(err)
			return
		}
		comics, err := m.store.ListComicsPerUserPSID(ctx, int32(userID))

		if len(comics) == 0 {
			sendTextBack(senderID, "Bạn chưa đăng ký nhận thông báo cho truyện nào")
			sendTextBack(senderID, "Nếu chưa biết cách đăng ký truyện hãy dùng lệnh /start và làm theo hướng dẫn ")
		} else {
			sendTextBack(senderID, fmt.Sprintf("Bạn đã đăng ký nhận thông báo cho %d truyện", len(comics)))
			sendTextBack(senderID, "Xem chi tiết tại www.cominify-bot.xyz")
		}
	case "/page":
		sendTextBack(senderID, `Các trạng hiện tại tôi hỗ trợ:\n
		beeng.net\n
		blogtruyen.vn\n
		truyenhtranh.net\n
		truyentranhtuan.com\n
		truyenqq.com\n
		hocvientruyentranh.net\n`)
	case "/tutor":
		sendTextBack(senderID, "Bạn có thể xem hướng dẫn tại: www.cominify-bot.xyz/tutorial hoặc dùng lệnh /start và làm theo hướng dẫn")
	default:
		sendTextBack(senderID, `Các lệnh tối hỗ trợ:
- /list:  xem các truyện đã đăng kí
- /page:  xem các trang web hiện tại BOT hỗ trợ
- /tutor: xem hướng dẫn`)
	}

	return
}

func (m *MSG) reponseGetStarted(ctx context.Context, senderID string) {

	sendTextBack(senderID, "Welcome to Comic Notify Bot!")
	sendTextBack(senderID, "Tôi là chatbot giúp theo dõi truyện tranh và thông báo mỗi khi truyện có chapter mới")
	sendTextBack(senderID, `Các lệnh tối hỗ trợ:
- /list:  xem các truyện đã đăng kí
- /page:  xem các trang web hiện tại BOT hỗ trợ
- /tutor: xem hướng dẫn`)

	m.responseCommand(ctx, senderID, "/start")
	return
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
		comic, err = m.crawler.GetComicInfo(ctx, comicURL)
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
