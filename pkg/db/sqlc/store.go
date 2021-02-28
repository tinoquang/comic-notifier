package db

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"strings"

	"github.com/tinoquang/comic-notifier/pkg/logging"
	"github.com/tinoquang/comic-notifier/pkg/util"
)

type Stores interface {
	Querier
	SubscribeComic(ctx context.Context, userPSID, comicURL string) (*Comic, error)
	CheckUserExist(ctx context.Context, userAppID string) error
	UpdateComicChapter(ctx context.Context, comic *Comic) error
	SynchronizedComicImage(comic *Comic) error
}

type crawler interface {
	GetComicInfo(ctx context.Context, comic *Comic) (err error)
	GetUserInfoFromFacebook(field, id string) (user User, err error)
	GetImg(comicPage, comicName string) error
	UploadImg(comicPage, comicName, imgURL string) (err error)
	DeleteImg(comicPage, comicName string) error
}

type StoreDB struct {
	db *sql.DB
	*Queries
	crawl crawler
}

// NewStore create new stores
func NewStore(dbconn *sql.DB, crawl crawler) Stores {
	return &StoreDB{
		db:      dbconn,
		Queries: New(dbconn),
		crawl:   crawl,
	}
}

func (s *StoreDB) execTx(ctx context.Context, fn func(Querier) error) error {

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		logging.Danger(err)
		return err
	}

	q := New(tx)
	err = fn(q)
	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			logging.Danger(rbErr)
			return fmt.Errorf("tx err: %v, rb err: %v", err, rbErr)
		}
		return err
	}

	return tx.Commit()
}

// SubscribeComic subscribe and return comic info to user
func (s *StoreDB) SubscribeComic(ctx context.Context, userPSID, comicURL string) (*Comic, error) {

	var comic Comic

	parsedURL, err := url.Parse(comicURL)
	if err != nil || parsedURL.Host == "" {
		return nil, util.ErrInvalidURL
	}

	err = s.execTx(ctx, func(q Querier) (txErr error) {

		// Check comic existed in DB, if not crawl comic's info
		comic, txErr = q.GetComicByURL(ctx, comicURL)
		if txErr != nil {

			comic.Page = parsedURL.Hostname()
			comic.Url = comicURL
			// Fail to get comic, return err
			if txErr != sql.ErrNoRows {
				logging.Danger(txErr)
				return
			}

			// Comic doesn't existed in DB --> crawl Comic info and add into DB
			txErr = s.crawl.GetComicInfo(ctx, &comic)
			if txErr != nil {
				logging.Danger(err)
				return
			}

			comic, txErr = s.CreateComic(ctx, CreateComicParams{
				Page:        comic.Page,
				Name:        comic.Name,
				Url:         comic.Url,
				ImgUrl:      comic.ImgUrl,
				CloudImgUrl: comic.CloudImgUrl,
				LatestChap:  comic.LatestChap,
				ChapUrl:     comic.ChapUrl,
			})
			if txErr != nil {
				logging.Danger(txErr)
				return
			}
		}

		// Check user existed in DB, if not crawl user's info
		user, txErr := s.GetUserByPSID(ctx, sql.NullString{String: userPSID, Valid: true})
		if txErr != nil {

			if txErr != sql.ErrNoRows {
				logging.Danger(txErr)
				return
			}

			user, txErr = s.crawl.GetUserInfoFromFacebook("psid", userPSID)
			if txErr != nil {
				logging.Danger(txErr)
				return
			}

			user, txErr = s.CreateUser(ctx, CreateUserParams{
				Name:       user.Name,
				Psid:       user.Psid,
				Appid:      user.Appid,
				ProfilePic: user.ProfilePic,
			})
			if txErr != nil {
				logging.Danger(err)
				return
			}
		}

		// Check comic already subscired
		_, txErr = s.GetSubscriber(ctx, GetSubscriberParams{UserID: user.ID, ComicID: comic.ID})
		if txErr == nil {
			return util.ErrAlreadySubscribed
		}

		if txErr != sql.ErrNoRows {
			logging.Danger(err)
			return
		}

		_, txErr = s.CreateSubscriber(ctx, CreateSubscriberParams{
			UserID:  user.ID,
			ComicID: comic.ID,
		})
		if txErr != nil {
			logging.Danger(err)
			return
		}

		// Last step, check comic's image in firebase DB
		txErr = s.crawl.GetImg(comic.Page, comic.Name)
		if txErr != nil {
			if strings.Contains(txErr.Error(), "object doesn't exist") {
				txErr = s.crawl.UploadImg(comic.Page, comic.Name, comic.ImgUrl)
				if txErr != nil {
					logging.Danger(txErr)
					return
				}
			} else {
				logging.Danger(txErr)
				return
			}

		}
		return nil
	})

	return &comic, err
}

// CheckUserExist call FacebookAPI to get userinfo, then update user DB if user already exist PSID
func (s *StoreDB) CheckUserExist(ctx context.Context, userAppID string) error {

	var user User
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
	user, err = s.crawl.GetUserInfoFromFacebook("appid", userAppID)
	if err != nil {
		logging.Danger(err)
		return err
	}

	// Check user's psid existed to
	if user.Psid.String != "" {
		_, err := s.GetUserByPSID(ctx, user.Psid)
		if err != nil {
			return err
		}

		_, err = s.UpdateUser(ctx, UpdateUserParams{
			Appid: user.Appid,
			Psid:  user.Psid,
		})

		if err != nil {
			logging.Danger(err)
			return err
		}
	}

	return nil
}

// UpdateComicChapter get comic info and compare to current comic in DB to verify new chapter release
func (s *StoreDB) UpdateComicChapter(ctx context.Context, comic *Comic) error {
	oldImgURL := comic.ImgUrl

	err := s.crawl.GetComicInfo(ctx, comic)
	if err != nil {
		return err
	}

	if oldImgURL != comic.ImgUrl {
		err = s.crawl.UploadImg(comic.Page, comic.Name, comic.ImgUrl)
		if err != nil {
			logging.Danger(err)
		}
	}

	_, err = s.UpdateComic(ctx, UpdateComicParams{
		ID:          comic.ID,
		LatestChap:  comic.LatestChap,
		ChapUrl:     comic.ChapUrl,
		ImgUrl:      comic.ImgUrl,
		CloudImgUrl: comic.CloudImgUrl,
	})
	if err != nil {
		return err
	}

	return nil
}

func (s *StoreDB) SynchronizedComicImage(comic *Comic) error {

	err := s.crawl.GetImg(comic.Page, comic.Name)
	if err == nil {
		return nil
	}

	if strings.Contains(err.Error(), "object doesn't exist") {
		err = s.crawl.UploadImg(comic.Page, comic.Name, comic.ImgUrl)
		if err != nil {
			return err
		}
	}

	return err
}
