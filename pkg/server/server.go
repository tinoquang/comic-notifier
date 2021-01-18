package server

import (
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

// New  create new server
func New(store db.Stores) *Server {

	// Get env config
	messengerEndpoint = conf.Cfg.Webhook.GraphEndpoint + "/me/messages"
	webhookToken = conf.Cfg.Webhook.WebhookToken
	pageToken = conf.Cfg.FBSecret.PakeToken

	s := &Server{
		API: NewAPI(store),
		Msg: NewMSG(store),
	}

	// Start update-comic thread
	go updateComicThread(store, conf.Cfg.WrkDat.WorkerNum, conf.Cfg.WrkDat.Timeout)
	return s
}
