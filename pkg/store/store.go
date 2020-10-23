package store

import (
	"context"
	"database/sql"
	"net/url"
	"strings"

	"github.com/pkg/errors"
	"github.com/tinoquang/comic-notifier/pkg/conf"
	"github.com/tinoquang/comic-notifier/pkg/db"
	"github.com/tinoquang/comic-notifier/pkg/model"
	"github.com/tinoquang/comic-notifier/pkg/server/crawler"
	"github.com/tinoquang/comic-notifier/pkg/server/img"
	"github.com/tinoquang/comic-notifier/pkg/util"
)

// Stores contain all store interfaces
type Stores struct {
	db         *sql.DB
	cfg        *conf.Config
	Comic      ComicInterface
	Page       PageInterface
	User       UserInterface
	Subscriber SubscriberInterface
}

// New create new stores
func New(db *sql.DB, cfg *conf.Config) *Stores {
	return &Stores{
		db:         db,
		cfg:        cfg,
		Comic:      NewComicStore(db, cfg),
		Page:       NewPageStore(db, cfg),
		User:       NewUserStore(db, cfg),
		Subscriber: NewSubscriberStore(db, cfg),
	}
}

// SubscribeComic subscribe and return comic info to user
func (s *Stores) SubscribeComic(ctx context.Context, field, id, comicURL string) (*model.Comic, error) {

	var err error
	var comic *model.Comic
	newComic := 1

	parsedURL, err := url.Parse(comicURL)
	if err != nil || parsedURL.Host == "" {
		return nil, errors.New("Please check your URL")
	}

	// Check page support, if not send back "Page is not supported"
	_, err = s.Page.GetByName(ctx, parsedURL.Hostname())
	if err != nil {
		return nil, errors.New("Sorry, page " + parsedURL.Hostname() + " is not supported yet")
	}

	err = db.WithTransaction(ctx, s.db, func(tx db.Transaction) (inErr error) {

		comic = &model.Comic{
			Page: parsedURL.Hostname(),
			URL:  comicURL,
		}
		// Get all comic infos includes latest chapter
		inErr = crawler.GetComicInfo(ctx, comic)
		if inErr != nil {
			util.Danger(inErr)
			return inErr
		}

		// Add new comic to DB
		query := `INSERT INTO comics (page, name, url, img_url, latest_chap, chap_url, date, date_format) 
					VALUES ($1, $2, $3, $4, $5, $6, $7, $8) 
					ON CONFLICT (url)
					DO NOTHING
					RETURNING id, imgur_id, imgur_link`
		inErr = tx.QueryRowContext(ctx, query, comic.Page, comic.Name, comic.URL, comic.ImageURL, comic.LatestChap, comic.ChapURL, comic.Date, comic.DateFormat).
			Scan(&comic.ID, &comic.ImgurID, &comic.ImgurLink)

		if inErr != nil {
			if inErr == sql.ErrNoRows {
				newComic = 0
			} else {
				util.Danger(inErr)
				return inErr
			}
		}

		// Validate users is in user DB or not
		// If not, add user to database, return "Subscribed to ..."
		// else return "Already subscribed"
		user, inErr := crawler.GetUserInfoByID(s.cfg, field, id)
		if inErr != nil {
			util.Danger(inErr)
			return
		}

		query = `INSERT INTO users (name, psid, appid, profile_pic) VALUES ($1, $2, $3, $4)
				ON CONFLICT (psid)
				DO NOTHING
				RETURNING psid`
		inErr = tx.QueryRowContext(ctx, query, user.Name, user.PSID, user.AppID, user.ProfilePic).Scan(&user.PSID)
		if inErr != nil && inErr != sql.ErrNoRows {
			util.Danger(inErr)
			return
		}

		subscriber, inErr := s.Subscriber.Get(ctx, user.PSID, comic.ID)
		if inErr != nil {
			if !strings.Contains(inErr.Error(), "not found") {
				util.Danger(inErr)
				return
			}

			// Add comic and user to subscribe table
			query = `INSERT INTO subscribers (user_psid, comic_id) VALUES ($1,$2) RETURNING id`
			inErr = tx.QueryRowContext(ctx, query, user.PSID, comic.ID).Scan(&subscriber.ID)
			if inErr != nil {
				return
			}
		} else {
			return errors.New("Already subscribed")
		}

		if newComic != 0 {
			image, e := img.UploadImagetoImgur(comic.Page+" "+comic.Name, comic.ImageURL)
			if e != nil {
				util.Danger(e)
				return e
			}
			comic.ImgurID = model.NullString(image.ID)
			comic.ImgurLink = model.NullString(image.Link)
			query := "UPDATE comics SET imgur_id=$2, imgur_link=$3 WHERE id=$1 RETURNING id, imgur_id, imgur_link"
			inErr = tx.QueryRowContext(ctx, query, comic.ID, image.ID, image.Link).Scan(&comic.ID, &comic.ImgurID, &comic.ImgurLink)
			if inErr != nil {
				img.DeleteImg(image.ID)
				return
			}
		}

		return
	})

	return comic, err
}
