package store

import (
	"database/sql"

	"github.com/tinoquang/comic-notifier/pkg/conf"
)

// Stores contain all store interfaces
type Stores struct {
	Comic ComicInterface
	Page  PageInterface
}

// New create new stores
func New(db *sql.DB, cfg *conf.Config) *Stores {
	return &Stores{
		Comic: NewComicStore(db, cfg),
		Page:  NewPageStore(db, cfg),
	}
}
