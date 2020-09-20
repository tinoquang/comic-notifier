package store

import (
	"database/sql"

	"github.com/tinoquang/comic-notifier/pkg/conf"
	"github.com/tinoquang/comic-notifier/pkg/util"
)

// ComicInterface contains function to interact with comic DB
type ComicInterface interface {
	GetByURL()
}

type comicDB struct {
	dbconn *sql.DB
	cfg    *conf.Config
}

// NewComicStore return comic interfaces
func NewComicStore(dbconn *sql.DB, cfg *conf.Config) ComicInterface {
	return &comicDB{dbconn: dbconn, cfg: cfg}
}

func (c *comicDB) GetByURL() {

	util.Info("Comic getbyURL")
}
