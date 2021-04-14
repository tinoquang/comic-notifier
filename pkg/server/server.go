package server

import (
	"context"
	"sync"

	"github.com/tinoquang/comic-notifier/pkg/conf"
	db "github.com/tinoquang/comic-notifier/pkg/db/sqlc"
)

// Server implement main business logic
type Server struct {
	API *API
	Msg *MSG
}

var (
	messengerEndpoint string
	pageToken         string
	webhookToken      string
)

// Crawler contain comic, user and image crawler
type infoCrawler interface {
	GetComicInfo(ctx context.Context, comicURL string, checkSpoiler bool) (comic db.Comic, err error)
	GetUserInfoFromFacebook(field, id string) (user db.User, err error)
}

// New  create new server
func New(store db.Store, crawler infoCrawler) *Server {

	// Get env config
	messengerEndpoint = conf.Cfg.Webhook.GraphEndpoint + "/me/messages"
	webhookToken = conf.Cfg.Webhook.WebhookToken
	pageToken = conf.Cfg.FBSecret.PakeToken

	s := &Server{
		API: NewAPI(store),
		Msg: NewMSG(store, crawler),
	}

	// Start update-comic
	updateLock := sync.Mutex{} // using lock to avoid updateService and notifyService run simuteneously

	initUpdateService(&updateLock, crawler, store, conf.Cfg.WrkDat.WorkerNum, conf.Cfg.WrkDat.Timeout)
	initNotifyService(&updateLock)
	return s
}
