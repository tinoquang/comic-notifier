package store

import (
	"context"
	"database/sql"

	"github.com/pkg/errors"
	"github.com/tinoquang/comic-notifier/pkg/conf"
	"github.com/tinoquang/comic-notifier/pkg/db"
	"github.com/tinoquang/comic-notifier/pkg/model"
	"github.com/tinoquang/comic-notifier/pkg/util"
)

// ComicInterface contains comic's interact method
type ComicInterface interface {
	GetByURL(ctx context.Context, URL string) (*model.Comic, error)
	Create(ctx context.Context, comic *model.Comic) error
	// Update()
	// List()
}

type comicDB struct {
	dbconn *sql.DB
	cfg    *conf.Config
}

// NewComicStore return comic interfaces
func NewComicStore(dbconn *sql.DB, cfg *conf.Config) ComicInterface {
	return &comicDB{dbconn: dbconn, cfg: cfg}
}

func (c *comicDB) GetByURL(ctx context.Context, URL string) (*model.Comic, error) {

	comics, err := c.getBySQL(ctx, "WHERE url=$1", URL)
	if err != nil {
		util.Danger()
		return nil, err
	}

	if len(comics) == 0 {
		return &model.Comic{}, errors.New("Comic not found")
	}

	return &comics[0], nil
}

func (c *comicDB) Create(ctx context.Context, comic *model.Comic) error {

	query := "INSERT INTO comics (name, url, image_url, latest_chap, chap_url, date, date_format) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id"

	err := db.WithTransaction(ctx, c.dbconn, func(tx db.Transaction) error {
		return tx.QueryRowContext(
			ctx, query, comic.Name, comic.URL, comic.ImageURL, comic.LatestChap, comic.ChapURL, comic.Date, comic.DateFormat,
		).Scan(&comic.ID)
	})
	return err

}

func (c *comicDB) getBySQL(ctx context.Context, query string, args ...interface{}) ([]model.Comic, error) {
	rows, err := c.dbconn.QueryContext(ctx, "SELECT * FROM comics "+query, args...)
	if err != nil {
		util.Danger()
		return nil, err
	}

	comics := []model.Comic{}
	defer rows.Close()
	for rows.Next() {
		comic := model.Comic{}
		err := rows.Scan(&comic.ID, &comic.Name, &comic.URL, &comic.ImageURL, &comic.LatestChap, &comic.ChapURL, &comic.Date, &comic.DateFormat)
		if err != nil {
			util.Danger()
			return nil, err
		}

		comics = append(comics, comic)
	}
	if err = rows.Err(); err != nil {
		util.Danger()
		return nil, err
	}

	return comics, nil
}
