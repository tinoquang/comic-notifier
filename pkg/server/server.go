package server

import (
	"github.com/tinoquang/comic-notifier/pkg/conf"
	"github.com/tinoquang/comic-notifier/pkg/server/crawler"
	"github.com/tinoquang/comic-notifier/pkg/server/img"
	"github.com/tinoquang/comic-notifier/pkg/store"
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

// New  create new server
func New(cfg *conf.Config, store *store.Stores) *Server {

	// Get env config
	messengerEndpoint = cfg.Webhook.GraphEndpoint + "/me/messages"
	webhookToken = cfg.Webhook.WebhookToken
	pageToken = cfg.FBSecret.PakeToken

	s := &Server{
		API: NewAPI(cfg, store),
		Msg: NewMSG(cfg, store),
	}

	crawler.New()
	img.SetEnvVar(cfg)

	// Start update-comic thread
	go updateComicThread(store, cfg.WrkDat.WorkerNum, cfg.WrkDat.Timeout)
	return s
}
