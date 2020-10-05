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

// Server implement main business logic
type Server struct {
	cfg   *conf.Config
	store *store.Stores
	comH  map[string]comicHandler
}

// New  create new server
func New(cfg *conf.Config, store *store.Stores) *Server {

	s := &Server{
		cfg:   cfg,
		store: store,
	}

	// Create map between comic page name and it's handler
	s.initComicHandler()

	return s
}

// GetPage (GET /pages/{name})
func (s *Server) GetPage(ctx context.Context, name string) (*model.Page, error) {

	return s.store.Page.GetByName(ctx, name)
}

/* ===================== User ============================ */

// Users (GET /users)
func (s *Server) Users(ctx context.Context) ([]model.User, error) {

	return s.store.User.List(ctx)
}

// GetUserByPSID (GET /users?psid=)
func (s *Server) GetUserByPSID(ctx context.Context, psid string) (*model.User, error) {
	return s.store.User.GetByFBID(ctx, "psid", psid)
}

// GetUserByComicID (GET /comics/{id}/users)
func (s *Server) GetUserByComicID(ctx context.Context, comicID int) ([]model.User, error) {

	return s.store.User.ListByComicID(ctx, comicID)
}

/* ===================== Comic ============================ */

// Comics (GET /comics)
func (s *Server) Comics(ctx context.Context) ([]model.Comic, error) {

	return s.store.Comic.List(ctx)
}

// UpdateComic use when new chapter realease
func (s *Server) UpdateComic(ctx context.Context, comic *model.Comic) (bool, error) {

	updated := true
	err := s.getComicInfo(ctx, comic)

	if err != nil {
		if strings.Contains(err.Error(), "No new chapter") {
			updated = false
		} else {
			util.Danger()
		}
		return updated, err
	}

	err = s.store.Comic.Update(ctx, comic)
	return updated, err
}

/* ===================== Subsciber ========================== */

// SubscribeComic (POST /users/{id}/comics)
func (s *Server) SubscribeComic(ctx context.Context, field string, id string, comicURL string) (*model.Comic, error) {

	parsedURL, err := url.Parse(comicURL)
	if err != nil || parsedURL.Host == "" {
		return nil, errors.New("Please check your URL")
	}

	// Check page support, if not send back "Page is not supported"
	_, err = s.store.Page.GetByName(ctx, parsedURL.Hostname())
	if err != nil {
		return nil, errors.New("Sorry, page " + parsedURL.Hostname() + " is not supported yet")
	}

	// Page URL validated, now check comics already in database
	// util.Info("Validated " + page.Name)
	comic, err := s.store.Comic.GetByURL(ctx, comicURL)

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
			err = s.getComicInfo(ctx, comic)
			if err != nil {
				util.Danger(err)
				return nil, errors.New("Please check your URL")
			}

			// Add new comic to DB
			err = s.store.Comic.Create(ctx, comic)
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
	user, err := s.store.User.GetByFBID(ctx, field, id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {

			util.Info("Add new user")

			user, err = s.getUserInfoByID(field, id)
			// Check user already exist
			if err != nil {
				util.Danger(err)
				return nil, errors.New("Please try again later")
			}
			err = s.store.User.Create(ctx, user)

			if err != nil {
				util.Danger(err)
				return nil, errors.New("Please try again later")
			}
		} else {
			util.Danger(err)
			return nil, errors.New("Please try again later")
		}
	}

	subscriber, err := s.store.Subscriber.Get(ctx, user.ID, comic.ID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			subscriber = &model.Subscriber{
				PSID:    user.PSID,
				ComicID: comic.ID,
			}

			err = s.store.Subscriber.Create(ctx, subscriber)
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
func (s *Server) UnsubscribeComic(ctx context.Context, psid string, comicID int) error {

	return s.store.Subscriber.Delete(ctx, userID, comicID)
}

// GetComicByUserID (GET /user/{user_id}/comic/{id})
func (s *Server) GetComicByUserID(ctx context.Context, psid string, comicID int) (*model.Comic, error) {
	_, err := s.store.Subscriber.Get(ctx, userID, comicID)
	if err != nil {
		return nil, err
	}

	return s.store.Comic.Get(ctx, comicID)
}
