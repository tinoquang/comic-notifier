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

// GetSubscriber (GET /subscriber/{id})
func (s *Server) GetSubscriber(ctx context.Context, id int) (*model.Subscriber, error) {
	return s.store.Subscriber.GetByID(ctx, id)
}

// SubscribeComic (POST /user/{id}/comics)
func (s *Server) SubscribeComic(ctx context.Context, field string, id string, comicURL string) (int, *model.Comic, error) {

	parsedURL, err := url.Parse(comicURL)
	if err != nil || parsedURL.Host == "" {
		return 0, nil, errors.New("Please check your URL")
	}

	// Check page support, if not send back "Page is not supported"
	_, err = s.store.Page.GetByName(ctx, parsedURL.Hostname())
	if err != nil {
		return 0, nil, errors.New("Sorry, page " + parsedURL.Hostname() + " is not supported yet")
	}

	// Page URL validated, now check comics already in database
	// util.Info("Validated " + page.Name)
	comic, err := s.store.Comic.GetByURL(ctx, comicURL)

	// If comic is not in database, query it's latest chap,
	// add to database, then prepare response with latest chapter
	if err != nil {
		if strings.Contains(err.Error(), "not found") {

			util.Info("Comic is not in DB yet, insert it")

			// Get all comic infos includes latest chapter
			comic, err = s.getComicInfo(ctx, parsedURL.Hostname(), comicURL)
			if err != nil {
				util.Danger(err)
				return 0, nil, errors.New("Please check your URL")
			}

			// Add new comic to DB
			err = s.store.Comic.Create(ctx, comic)
			if err != nil {
				util.Danger(err)
				return 0, nil, errors.New("Please try again later")
			}
		} else {
			util.Danger(err)
			return 0, nil, errors.New("Please try again later")
		}
	}

	// Validate users is in user DB or not
	// If not, add user to database, return "Subscribed to ..."
	// else return "Already subscribed"
	user, err := s.store.User.GetByID(ctx, field, id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {

			util.Info("Add new user")

			user, err = s.getUserInfoByID(field, id)
			// Check user already exist
			if err != nil {
				util.Danger(err)
				return 0, nil, errors.New("Please try again later")
			}
			err = s.store.User.Create(ctx, user)

			if err != nil {
				util.Danger(err)
				return 0, nil, errors.New("Please try again later")
			}
		} else {
			util.Danger(err)
			return 0, nil, errors.New("Please try again later")
		}
	}

	subscriber, err := s.store.Subscriber.Get(ctx, user.ID, comic.ID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			subscriber = &model.Subscriber{
				Page:      comic.Page,
				UserID:    user.ID,
				UserName:  user.Name,
				ComicID:   comic.ID,
				ComicName: comic.Name,
			}

			err = s.store.Subscriber.Create(ctx, subscriber)
			if err != nil {
				util.Danger(err)
				return 0, nil, errors.New("Please try again later")
			}
			return subscriber.ID, comic, nil
		}
		util.Danger(err)
		return 0, nil, errors.New("Please try again later")
	}
	return 0, nil, errors.New("Already subscribed")
}
