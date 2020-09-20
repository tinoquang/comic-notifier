package store

import (
	"database/sql"

	"github.com/tinoquang/comic-notifier/pkg/conf"
)

type SubscriberInterface interface {
}

type subscriberDB struct {
	dbconn *sql.DB
	cfb    *conf.Config
}
