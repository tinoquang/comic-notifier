package store

import (
	"database/sql"

	"github.com/tinoquang/comic-notifier/pkg/conf"
)

// Stores contain all store interfaces
type Stores struct {
	Comic      ComicInterface
	Page       PageInterface
	User       UserInterface
	Subscriber SubscriberInterface
}

// New create new stores
func New(db *sql.DB, cfg *conf.Config) *Stores {
	return &Stores{
		Comic:      NewComicStore(db, cfg),
		Page:       NewPageStore(db, cfg),
		User:       NewUserStore(db, cfg),
		Subscriber: NewSubscriberStore(db, cfg),
	}
}
