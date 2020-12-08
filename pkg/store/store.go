package store

import (
	"context"
	"database/sql"
	"errors"
	"net/url"
	"strings"

	"github.com/keegancsmith/sqlf"
	"github.com/tinoquang/comic-notifier/pkg/db"
	"github.com/tinoquang/comic-notifier/pkg/logging"
	"github.com/tinoquang/comic-notifier/pkg/model"
	"github.com/tinoquang/comic-notifier/pkg/server/crawler"
	"github.com/tinoquang/comic-notifier/pkg/server/img"
	"github.com/tinoquang/comic-notifier/pkg/util"
)

var (
	ErrAlreadySubscribed = errors.New("Already subscribed")
	ErrNotFound          = errors.New("Not found")
)

// Stores contain all store interfaces
type Stores struct {
	db         *sql.DB
	Comic      ComicInterface
	User       UserInterface
	Subscriber SubscriberInterface
}

// New create new stores
func New(db *sql.DB) *Stores {
	return &Stores{
		db:         db,
		Comic:      NewComicStore(db),
		User:       NewUserStore(db),
		Subscriber: NewSubscriberStore(db),
	}
}

// SubscribeComic subscribe and return comic info to user
func (s *Stores) SubscribeComic(ctx context.Context, userPSID, comicURL string) (*model.Comic, error) {

	var err error
	var comic *model.Comic
	newComic := 1

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
			inErr = crawler.GetComicInfo(ctx, comic)
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
		} else {
			newComic = 0
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
			if inErr != ErrNotFound {
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
			return ErrAlreadySubscribed
		}

		if newComic != 0 {
			e := comic.UpdateCloudImg()
			if e != nil {
				logging.Danger(e)
				return e
			}

			query := "UPDATE comics SET cloud_img=$2 WHERE id=$1 RETURNING id, cloud_img"
			inErr = tx.QueryRowContext(ctx, query, comic.ID, comic.CloudImg).Scan(&comic.ID, &comic.CloudImg)
			if inErr != nil {
				logging.Danger(inErr)
				img.DeleteFirebaseImg(comic.Page, comic.Name)
				return
			}

		}

		return
	})

	return comic, err
}

// LimitOffset specifies SQL LIMIT and OFFSET counts. A pointer to it is typically embedded in other options
// structures that need to perform SQL queries with LIMIT and OFFSET.
type LimitOffset struct {
	Limit  int // SQL LIMIT count
	Offset int // SQL OFFSET count
}

// SQL returns the SQL query fragment ("LIMIT %d OFFSET %d") for use in SQL queries.
func (o *LimitOffset) SQL() *sqlf.Query {
	if o == nil {
		return &sqlf.Query{}
	}

	if o.Limit == 0 {
		return sqlf.Sprintf("LIMIT ALL OFFSET %d", o.Offset)
	}

	return sqlf.Sprintf("LIMIT %d OFFSET %d", o.Limit, o.Offset)
}

// NameLikeOptions used to query by name using like
type NameLikeOptions struct {
	// Query specifies a search query for organizations.
	Query string
}

// ListComicNameLikeSQL used to search by name if query is set
func ListComicNameLikeSQL(opt *NameLikeOptions) (conds []*sqlf.Query) {
	conds = []*sqlf.Query{sqlf.Sprintf("TRUE")}
	if opt.Query != "" {
		query := "%" + strings.Replace(opt.Query, " ", "%", -1) + "%"
		conds = append(conds, sqlf.Sprintf("comics.name ILIKE %s or unaccent(comics.name) ILIKE %s", query, query))
	}
	return conds
}
