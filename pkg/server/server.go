package server

import (
	"context"

	"github.com/PuerkitoBio/goquery"
	"github.com/tinoquang/comic-notifier/pkg/conf"
	"github.com/tinoquang/comic-notifier/pkg/model"
	"github.com/tinoquang/comic-notifier/pkg/store"
)

// Server implement main business logic
type Server struct {
	api *API
	msg *MSG
}

type comicHandler func(ctx context.Context, doc *goquery.Document, comic *model.Comic) (err error)

var handler map[string]comicHandler

// New  create new server
func New(cfg *conf.Config, store *store.Stores) *Server {

	s := &Server{
		api: NewAPI(cfg, store),
		msg: NewMSG(cfg, store),
	}

	// Create map between comic page name and it's handler
	initComicHandler()

	return s
}

func initComicHandler() {

	handler = make(map[string]comicHandler)

	handler["beeng.net"] = handleBeeng
	handler["mangak.info"] = handleMangaK
	handler["truyenqq.com"] = handleTruyenqq
	handler["blogtruyen.vn"] = handleBlogTruyen

}
