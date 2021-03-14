package db

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/tinoquang/comic-notifier/pkg/logging"
)

type Store interface {
	Querier
	SubscribeComic(ctx context.Context, comic *Comic, user *User) error
	UpdateComicChapter(ctx context.Context, comic *Comic, oldImgURL string) (err error)
	SynchronizedComicImage(comic *Comic) error
	RemoveComic(ctx context.Context, comicID int32) error
}

type store struct {
	db *sql.DB
	*Queries
	firebase *firebaseConnection
}

// NewStore create new stores
func NewStore(dbconn *sql.DB) Store {
	return &store{
		db:       dbconn,
		Queries:  New(dbconn),
		firebase: newFirebaseConnection(),
	}
}

func (s *store) execTx(ctx context.Context, fn func(Querier) error) error {

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
func (s *store) SubscribeComic(ctx context.Context, comic *Comic, user *User) error {

	err := s.execTx(ctx, func(q Querier) (txErr error) {

		// Check comic existed in DB, if not crawl comic's info
		c, txErr := q.CreateComic(ctx, CreateComicParams{
			Page:        comic.Page,
			Name:        comic.Name,
			Url:         comic.Url,
			ImgUrl:      comic.ImgUrl,
			CloudImgUrl: comic.CloudImgUrl,
			LatestChap:  comic.LatestChap,
			ChapUrl:     comic.ChapUrl,
		})
		if txErr != nil && txErr != sql.ErrNoRows {
			logging.Danger(txErr)
			return
		}

		if c.ID == 0 {
			// Duplicate record
			c.ID = comic.ID
		}

		u, txErr := q.CreateUser(ctx, CreateUserParams{
			Name:       user.Name,
			Psid:       user.Psid,
			Appid:      user.Appid,
			ProfilePic: user.ProfilePic,
		})
		if txErr != nil && txErr != sql.ErrNoRows {
			logging.Danger(txErr)
			return
		}

		if u.ID == 0 {
			// Duplicate record
			u.ID = user.ID
		}

		logging.Info(comic)
		logging.Info(user)
		_, txErr = q.CreateSubscriber(ctx, CreateSubscriberParams{
			UserID:  u.ID,
			ComicID: c.ID,
		})
		if txErr != nil {
			logging.Danger(txErr)
			return
		}

		// Last step, check comic's image in firebase DB
		txErr = s.firebase.GetImg(comic.Page, comic.Name)
		if txErr != nil {
			if strings.Contains(txErr.Error(), "object doesn't exist") {
				txErr = s.firebase.UploadImg(comic.Page, comic.Name, comic.ImgUrl)
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

	return err
}

// UpdateComicChapter get comic info and compare to current comic in DB to verify new chapter release
func (s *store) UpdateComicChapter(ctx context.Context, comic *Comic, oldImgURL string) (err error) {

	if oldImgURL != comic.ImgUrl {
		err = s.firebase.UploadImg(comic.Page, comic.Name, comic.ImgUrl)
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

// SynchronizedComicImage check comic's image exists in Firebase and sync with comic in DB
func (s *store) SynchronizedComicImage(comic *Comic) error {

	err := s.firebase.GetImg(comic.Page, comic.Name)
	if err == nil {
		return nil
	}

	if strings.Contains(err.Error(), "object doesn't exist") {
		err = s.firebase.UploadImg(comic.Page, comic.Name, comic.ImgUrl)
		if err != nil {
			return err
		}
	}

	return err
}

// RemoveComic delete comic in DB and image in firebase
func (s *store) RemoveComic(ctx context.Context, comicID int32) error {

	comic, err := s.GetComic(ctx, comicID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		logging.Danger(err)
		return err
	}

	err = s.DeleteComic(ctx, comicID)
	if err != nil {
		logging.Danger(err)
		return err
	}

	err = s.firebase.DeleteImg(comic.Page, comic.Name)
	if err != nil {
		logging.Danger(err)
		return err
	}

	return nil
}
