package store

import (
	"context"
	"database/sql"
	"net/url"

	"github.com/tinoquang/comic-notifier/pkg/db"
	"github.com/tinoquang/comic-notifier/pkg/logging"
	"github.com/tinoquang/comic-notifier/pkg/model"

	// "github.com/tinoquang/comic-notifier/pkg/model"
	"github.com/tinoquang/comic-notifier/pkg/util"
)

// Stores contain all store interfaces
type Stores struct {
	db         *sql.DB
	Comic      ComicRepo
	User       UserRepo
	Subscriber SubscriberRepo
}

// New create new stores
func New(db *sql.DB, firebaseDB *db.FirebaseDB) *Stores {
	return &Stores{
		db:         db,
		Comic:      newComicRepo(db, firebaseDB),
		User:       newUserRepo(db),
		Subscriber: newSubscriberRepo(db),
	}
}

// SubscribeComic subscribe and return comic info to user
func (s *Stores) SubscribeComic(ctx context.Context, userPSID, comicURL string) (*model.Comic, error) {

	var err error
	var comic *model.Comic

	parsedURL, err := url.Parse(comicURL)
	if err != nil || parsedURL.Host == "" {
		return nil, util.ErrInvalidURL
	}

	err = db.WithTransaction(ctx, s.db, func(tx db.Transaction) (inErr error) {

		comic = &model.Comic{
			Page: parsedURL.Hostname(),
			URL:  comicURL,
		}

		inErr = tx.QueryRowContext(ctx, "SELECT * from comics WHERE url=$1", comicURL).
			Scan(&comic.ID, &comic.Page, &comic.Name, &comic.URL, &comic.OriginImgURL, &comic.CloudImg, &comic.LatestChap, &comic.ChapURL)
		if inErr != nil {

			if inErr != sql.ErrNoRows {
				logging.Danger(inErr)
				return inErr
			}

			// Get all comic infos includes latest chapter
			// inErr = crawler.GetComicInfo(ctx, comic)
			if inErr != nil {
				// logging.Danger(inErr)
				return inErr
			}
			// Add new comic to DB
			query := `INSERT INTO comics (page, name, url, img_url, latest_chap, chap_url) 
						VALUES ($1, $2, $3, $4, $5, $6) 
						RETURNING id`
			inErr = tx.QueryRowContext(ctx, query, comic.Page, comic.Name, comic.URL, comic.OriginImgURL, comic.LatestChap, comic.ChapURL).
				Scan(&comic.ID)

			if inErr != nil {
				logging.Danger(inErr)
				return inErr
			}
		}

		// Validate users is in user DB or not
		// If not, add user to database, return "Subscribed to ..."
		// else return "Already subscribed"
		user := &model.User{}
		inErr = tx.QueryRowContext(ctx, "SELECT * from users WHERE psid=$1", userPSID).Scan(&user.Name, &user.PSID, &user.AppID, &user.ProfilePic)
		if inErr != nil {

			if inErr != sql.ErrNoRows {
				logging.Danger(inErr)
				return inErr
			}

			inErr = user.GetInfoFromFB("psid", userPSID)
			if inErr != nil {
				logging.Danger(inErr)
				return inErr
			}

			query := `INSERT INTO users (name, psid, appid, profile_pic) VALUES ($1, $2, $3, $4) RETURNING psid`
			inErr = tx.QueryRowContext(ctx, query, user.Name, user.PSID, user.AppID, user.ProfilePic).Scan(&user.PSID)
			if inErr != nil && inErr != sql.ErrNoRows {
				logging.Danger(inErr)
				return inErr
			}
		}

		subscriber, inErr := s.Subscriber.Get(ctx, user.PSID, comic.ID)
		if inErr != nil {
			if inErr != util.ErrNotFound {
				logging.Danger(inErr)
				return
			}

			// Add comic and user to subscribe table
			query := `INSERT INTO subscribers (user_psid, comic_id) VALUES ($1,$2) RETURNING id`
			inErr = tx.QueryRowContext(ctx, query, user.PSID, comic.ID).Scan(&subscriber.ID)
			if inErr != nil {
				return
			}
		} else {
			return util.ErrAlreadySubscribed
		}

		return
	})

	return comic, err
}
