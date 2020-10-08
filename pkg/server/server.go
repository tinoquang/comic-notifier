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
	API *API
	Msg *MSG
}

type comicHandler func(ctx context.Context, doc *goquery.Document, comic *model.Comic) (err error)

var (
	messengerEndpoint string
	pageToken         string
	webhookToken      string
	handler           map[string]comicHandler
)

// New  create new server
func New(cfg *conf.Config, store *store.Stores) *Server {

	// Get env config
	messengerEndpoint = cfg.Webhook.GraphEndpoint + "me/messages"
	webhookToken = cfg.Webhook.WebhookToken
	pageToken = cfg.FBSecret.PakeToken

	s := &Server{
		API: NewAPI(cfg, store),
		Msg: NewMSG(cfg, store),
	}

	// Create map between comic page name and it's handler
	initComicHandler()

	// Start update-comic thread
	go updateComicThread(store, cfg.WrkDat.WorkerNum, cfg.WrkDat.Timeout)
	return s
}

func initComicHandler() {

	handler = make(map[string]comicHandler)

	handler["beeng.net"] = handleBeeng
	handler["mangak.info"] = handleMangaK
	handler["truyenqq.com"] = handleTruyenqq
	handler["blogtruyen.vn"] = handleBlogTruyen

}
