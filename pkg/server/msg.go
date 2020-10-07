package server

import (
	"context"
	"net/url"
	"strings"

	"github.com/pkg/errors"
	"github.com/tinoquang/comic-notifier/pkg/conf"
	"github.com/tinoquang/comic-notifier/pkg/model"
	"github.com/tinoquang/comic-notifier/pkg/store"
	"github.com/tinoquang/comic-notifier/pkg/util"
)

// MSG -> server handler for messenger endpoint
type MSG struct {
	cfg   *conf.Config
	store *store.Stores
}

// NewMSG return new api interface
func NewMSG(c *conf.Config, s *store.Stores) *MSG {
	return &MSG{cfg: c, store: s}
}

// GetPage (GET /pages/{name})
func (m *MSG) GetPage(ctx context.Context, name string) (*model.Page, error) {

	return m.store.Page.GetByName(ctx, name)
}

// Comics (GET /comics)
func (m *MSG) Comics(ctx context.Context) ([]model.Comic, error) {

	return m.store.Comic.List(ctx)
}

/* ===================== User ============================ */

// Users (GET /users)
func (m *MSG) Users(ctx context.Context) ([]model.User, error) {

	return m.store.User.List(ctx)
}

// GetUserByPSID (GET /users?psid=)
func (m *MSG) GetUserByPSID(ctx context.Context, psid string) (*model.User, error) {
	return m.store.User.GetByFBID(ctx, "psid", psid)
}

// GetUsersByComicID (GET /comics/{id}/users)
func (m *MSG) GetUsersByComicID(ctx context.Context, comicID int) ([]model.User, error) {

	return m.store.User.ListByComicID(ctx, comicID)
}

/* ===================== Comic ============================ */

// UpdateComic use when new chapter realease
func (m *MSG) UpdateComic(ctx context.Context, comic *model.Comic) (bool, error) {

	updated := true
	err := getComicInfo(ctx, comic)

	if err != nil {
		if strings.Contains(err.Error(), "No new chapter") {
			updated = false
		} else {
			util.Danger()
		}
		return updated, err
	}

	err = m.store.Comic.Update(ctx, comic)
	return updated, err
}

// SubscribeComic (POST /users/{id}/comics)
func (m *MSG) SubscribeComic(ctx context.Context, field string, id string, comicURL string) (*model.Comic, error) {

	parsedURL, err := url.Parse(comicURL)
	if err != nil || parsedURL.Host == "" {
		return nil, errors.New("Please check your URL")
	}

	// Check page support, if not send back "Page is not supported"
	_, err = m.store.Page.GetByName(ctx, parsedURL.Hostname())
	if err != nil {
		return nil, errors.New("Sorry, page " + parsedURL.Hostname() + " is not supported yet")
	}

	// Page URL validated, now check comics already in database
	// util.Info("Validated " + page.Name)
	comic, err := m.store.Comic.GetByURL(ctx, comicURL)

	// If comic is not in database, query it's latest chap,
	// add to database, then prepare response with latest chapter
	if err != nil {
		if strings.Contains(err.Error(), "not found") {

			util.Info("Comic is not in DB yet, insert it")
			comic = &model.Comic{
				Page: parsedURL.Hostname(),
				URL:  comicURL,
			}
			// Get all comic infos includes latest chapter
			err = getComicInfo(ctx, comic)
			if err != nil {
				util.Danger(err)
				return nil, errors.New("Please check your URL")
			}

			// Add new comic to DB
			err = m.store.Comic.Create(ctx, comic)
			if err != nil {
				util.Danger(err)
				return nil, errors.New("Please try again later")
			}
		} else {
			util.Danger(err)
			return nil, errors.New("Please try again later")
		}
	}

	// Validate users is in user DB or not
	// If not, add user to database, return "Subscribed to ..."
	// else return "Already subscribed"
	user, err := m.store.User.GetByFBID(ctx, field, id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {

			util.Info("Add new user")

			user, err = getUserInfoByID(field, id, m.cfg)
			// Check user already exist
			if err != nil {
				util.Danger(err)
				return nil, errors.New("Please try again later")
			}
			err = m.store.User.Create(ctx, user)

			if err != nil {
				util.Danger(err)
				return nil, errors.New("Please try again later")
			}
		} else {
			util.Danger(err)
			return nil, errors.New("Please try again later")
		}
	}

	_, err = m.store.Comic.GetByPSID(ctx, user.PSID, comic.ID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			subscriber := &model.Subscriber{
				PSID:    user.PSID,
				ComicID: comic.ID,
			}

			err = m.store.Subscriber.Create(ctx, subscriber)
			if err != nil {
				util.Danger(err)
				return nil, errors.New("Please try again later")
			}
			return comic, nil
		}
		util.Danger(err)
		return nil, errors.New("Please try again later")
	}
	return nil, errors.New("Already subscribed")
}

// UnsubscribeComic (DELETE /user/{user_id}/comic/{id})
func (m *MSG) UnsubscribeComic(ctx context.Context, psid string, comicID int) error {

	return m.store.Subscriber.Delete(ctx, psid, comicID)
}

// GetUserComic (GET /user/{user_id}/comics/{id})
func (m *MSG) GetUserComic(ctx context.Context, psid string, comicID int) (*model.Comic, error) {

	return m.store.Comic.GetByPSID(ctx, psid, comicID)
}

// GetUserComics (GET /user/{user_id}/comics)
func (m *MSG) GetUserComics(ctx context.Context, psid string) ([]model.Comic, error) {

	return m.store.Comic.ListByPSID(ctx, psid)
}
