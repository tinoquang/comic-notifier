package store

import (
	"context"
	"database/sql"

	"github.com/tinoquang/comic-notifier/pkg/conf"
	"github.com/tinoquang/comic-notifier/pkg/model"
	"github.com/tinoquang/comic-notifier/pkg/util"
)

// ComicInterface contains comic's interact method
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

func (c *comicDB) getBySQL(ctx context.Context, query string, args ...interface{}) ([]model.Comic, error) {
	rows, err := c.dbconn.QueryContext(ctx, "SELECT * FROM comics "+query, args...)
	if err != nil {
		return nil, err
	}

	comics := []model.Comic{}
	defer rows.Close()
	for rows.Next() {
		comic := model.Comic{}
		err := rows.Scan(&comic.ID, &comic.Name, &comic.URL, &comic.ImageURL, &comic.LatestChap, &comic.ChapURL, &comic.Date, &comic.DateFormat)
		if err != nil {
			return nil, err
		}

		comics = append(comics, comic)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return comics, nil
}
