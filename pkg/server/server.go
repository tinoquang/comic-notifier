package server

import (
	"context"

	"github.com/tinoquang/comic-notifier/pkg/conf"
	"github.com/tinoquang/comic-notifier/pkg/model"
	"github.com/tinoquang/comic-notifier/pkg/store"
)

// SvrInterface contain all server's method
type SvrInterface interface {
	GetPage(ctx context.Context, name string) (*model.Page, error)
}

// Server implement main business logic
type Server struct {
	cfg   *conf.Config
	store *store.Stores
}

// New  create new server
func New(cfg *conf.Config, store *store.Stores) SvrInterface {
	return &Server{
		cfg:   cfg,
		store: store,
	}
}

// GetPage (GET /pages/{name})
func (s *Server) GetPage(ctx context.Context, name string) (*model.Page, error) {

	return s.store.Page.GetByName(ctx, name)
}
