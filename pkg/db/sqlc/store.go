package db

import (
	"context"
	"database/sql"
	"net/url"

	"github.com/tinoquang/comic-notifier/pkg/logging"
	"github.com/tinoquang/comic-notifier/pkg/util"
)

type Stores interface {
	Querier
	SubscribeComic(ctx context.Context, userPSID, comicURL string) (*Comic, error)
	CheckUserExist(ctx context.Context, userAppID string) error
}

type crawler interface {
	GetComicInfo(ctx context.Context, comic *Comic) (err error)
	GetUserInfoFromFacebook(field, id string, user *CreateUserParams) error
	GetImg(comicPage, comicName string) error
	UploadImg(comicPage, comicName, imgURL string) (err error)
	DeleteImg(comicPage, comicName string) error
}

type StoreDB struct {
	db *sql.DB
	*Queries
	crawl crawler
}

// New create new stores
func NewStore(dbconn *sql.DB, crawl crawler) Stores {
	return &StoreDB{
		db:      dbconn,
		Queries: New(dbconn),
		crawl:   crawl,
	}
}

// SubscribeComic subscribe and return comic info to user
func (s *StoreDB) SubscribeComic(ctx context.Context, userPSID, comicURL string) (*Comic, error) {

	var err error
	var comic *Comic

	parsedURL, err := url.Parse(comicURL)
	if err != nil || parsedURL.Host == "" {
		return nil, util.ErrInvalidURL
	}

	// err = db.WithTransaction(ctx, s.db, func(tx db.Transaction) (inErr error) {

	// 	comic = &model.Comic{
	// 		Page: parsedURL.Hostname(),
	// 		URL:  comicURL,
	// 	}

	// 	inErr = tx.QueryRowContext(ctx, "SELECT * from comics WHERE url=$1", comicURL).
	// 		Scan(&comic.ID, &comic.Page, &comic.Name, &comic.URL, &comic.OriginImgURL, &comic.CloudImg, &comic.LatestChap, &comic.ChapURL)
	// 	if inErr != nil {

	// 		if inErr != sql.ErrNoRows {
	// 			logging.Danger(inErr)
	// 			return inErr
	// 		}

	// 		// Get all comic infos includes latest chapter
	// 		// inErr = crawler.GetComicInfo(ctx, comic)
	// 		if inErr != nil {
	// 			// logging.Danger(inErr)
	// 			return inErr
	// 		}
	// 		// Add new comic to DB
	// 		query := `INSERT INTO comics (page, name, url, img_url, latest_chap, chap_url)
	// 					VALUES ($1, $2, $3, $4, $5, $6)
	// 					RETURNING id`
	// 		inErr = tx.QueryRowContext(ctx, query, comic.Page, comic.Name, comic.URL, comic.OriginImgURL, comic.LatestChap, comic.ChapURL).
	// 			Scan(&comic.ID)

	// 		if inErr != nil {
	// 			logging.Danger(inErr)
	// 			return inErr
	// 		}
	// 	}

	// 	// Validate users is in user DB or not
	// 	// If not, add user to database, return "Subscribed to ..."
	// 	// else return "Already subscribed"
	// 	user := &model.User{}
	// 	inErr = tx.QueryRowContext(ctx, "SELECT * from users WHERE psid=$1", userPSID).Scan(&user.Name, &user.PSID, &user.AppID, &user.ProfilePic)
	// 	if inErr != nil {

	// 		if inErr != sql.ErrNoRows {
	// 			logging.Danger(inErr)
	// 			return inErr
	// 		}

	// 		inErr = user.GetInfoFromFB("psid", userPSID)
	// 		if inErr != nil {
	// 			logging.Danger(inErr)
	// 			return inErr
	// 		}

	// 		query := `INSERT INTO users (name, psid, appid, profile_pic) VALUES ($1, $2, $3, $4) RETURNING psid`
	// 		inErr = tx.QueryRowContext(ctx, query, user.Name, user.PSID, user.AppID, user.ProfilePic).Scan(&user.PSID)
	// 		if inErr != nil && inErr != sql.ErrNoRows {
	// 			logging.Danger(inErr)
	// 			return inErr
	// 		}
	// 	}

	// 	subscriber, inErr := s.Subscriber.Get(ctx, user.PSID, comic.ID)
	// 	if inErr != nil {
	// 		if inErr != util.ErrNotFound {
	// 			logging.Danger(inErr)
	// 			return
	// 		}

	// 		// Add comic and user to subscribe table
	// 		query := `INSERT INTO subscribers (user_psid, comic_id) VALUES ($1,$2) RETURNING id`
	// 		inErr = tx.QueryRowContext(ctx, query, user.PSID, comic.ID).Scan(&subscriber.ID)
	// 		if inErr != nil {
	// 			return
	// 		}
	// 	} else {
	// 		return util.ErrAlreadySubscribed
	// 	}

	// 	return
	// })

	return comic, err
}

// CheckUserExist call FacebookAPI to get userinfo, then update user DB if user already exist PSID
func (s *StoreDB) CheckUserExist(ctx context.Context, userAppID string) error {

	var userParam CreateUserParams

	// Check user already existed in DB
	_, err := s.GetUserByAppID(ctx, sql.NullString{
		String: userAppID,
		Valid:  true,
	})

	if err == nil {
		return nil
	}

	if err != sql.ErrNoRows {
		logging.Danger(err)
		return err
	}

	// So user which appid = userAppID is not existed in DB
	// we need to verify if same user existed by check user's psid

	// Get user info first
	err = s.crawl.GetUserInfoFromFacebook("appid", userAppID, &userParam)
	if err != nil {
		logging.Danger(err)
		return err
	}

	// Check user's psid existed to
	if userParam.Psid.String != "" {
		_, err := s.GetUserByPSID(ctx, userParam.Psid)
		if err != nil {
			logging.Danger(err)
			return err
		}

		_, err = s.UpdateUser(ctx, UpdateUserParams{
			Appid: userParam.Appid,
			Psid:  userParam.Psid,
		})

		if err != nil {
			logging.Danger(err)
			return err
		}
	}

	return nil
}
