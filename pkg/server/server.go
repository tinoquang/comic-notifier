package server

import (
	"context"
	"fmt"
	"net/url"

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

// SubscribeComic (POST /user/{id}/comics)
func (s *Server) SubscribeComic(ctx context.Context, comicURL string) (*model.Comic, error) {

	parsedURL, err := url.Parse(comicURL)
	if err != nil || parsedURL.Host == "" {
		return nil, errors.New("Please check your URL")
	}

	// Check page support, if not send back "Page is not supported"
	page, err := s.store.Page.GetByName(ctx, parsedURL.Hostname())
	if err != nil {
		return nil, errors.New("Sorry, page " + parsedURL.Hostname() + " is not supported yet")
	}

	fmt.Println(page)

	// Page URL validated, now check comics already in database
	util.Info("Validated " + page.Name)
	_, err = s.store.Comic.GetByURL(ctx, comicURL)

	// If comic is not in database, query it's latest chap,
	// add to database, then prepare response with latest chapter URL
	if err != nil {

		util.Info("Comic is not in DB yet, insert it")
		comic := &model.Comic{
			URL: comicURL,
		}
		err := s.getLatestChapter(ctx, parsedURL.Hostname(), comic)
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

		fmt.Printf("%+v\n", comic)
	}

	// err = data.AddComic(&comic)
	// if err != nil {
	// 	util.Danger(err)
	// 	sendTextBack(m.Sender.ID, "Try again later !")
	// 	return
	// }

	return new(model.Comic), nil
}
