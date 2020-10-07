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
	Get(ctx context.Context, id int) (*model.Comic, error)
	GetByURL(ctx context.Context, URL string) (*model.Comic, error)
	GetByPSID(ctx context.Context, psid string, comicID int) (*model.Comic, error)
	Create(ctx context.Context, comic *model.Comic) error
	Update(ctx context.Context, comic *model.Comic) error
	List(ctx context.Context) ([]model.Comic, error)
	ListByPSID(ctx context.Context, psid string) ([]model.Comic, error)
}

type comicDB struct {
	dbconn *sql.DB
	cfg    *conf.Config
}

// NewComicStore return comic interfaces
func NewComicStore(dbconn *sql.DB, cfg *conf.Config) ComicInterface {
	return &comicDB{dbconn: dbconn, cfg: cfg}
}

func (c *comicDB) Get(ctx context.Context, id int) (*model.Comic, error) {

	comics, err := c.getBySQL(ctx, "WHERE id=$1", id)
	if err != nil {
		util.Danger()
		return nil, err
	}

	if len(comics) == 0 {
		return &model.Comic{}, errors.New("Comic not found")
	}

	return &comics[0], nil
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

func (c *comicDB) GetByPSID(ctx context.Context, psid string, comicID int) (*model.Comic, error) {

	query := "LEFT JOIN subscribers ON comics.id=subscribers.comic_id && subscibers.psid=$1"

	comics, err := c.getBySQL(ctx, query, psid)
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

	query := "INSERT INTO comics (page, name, url, image_url, latest_chap, chap_url, date, date_format) VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id"

	err := db.WithTransaction(ctx, c.dbconn, func(tx db.Transaction) error {
		return tx.QueryRowContext(
			ctx, query, comic.Page, comic.Name, comic.URL, comic.ImageURL, comic.LatestChap, comic.ChapURL, comic.Date, comic.DateFormat,
		).Scan(&comic.ID)
	})
	return err

}

func (c *comicDB) Update(ctx context.Context, comic *model.Comic) error {

	query := "UPDATE comics SET latest_chap=$2, chap_url=$3, date=$4 WHERE id=$1"
	_, err := c.dbconn.ExecContext(ctx, query, comic.ID, comic.LatestChap, comic.ChapURL, comic.Date)
	return err
}

func (c *comicDB) List(ctx context.Context) ([]model.Comic, error) {

	return c.getBySQL(ctx, "")
}

func (c *comicDB) ListByPSID(ctx context.Context, psid string) ([]model.Comic, error) {
	query := "LEFT JOIN subscribers ON comics.id=subscribers.comic_id && subscibers.psid=$1"

	comics, err := c.getBySQL(ctx, query, psid)
	if err != nil || len(comics) == 0 {
		util.Danger()
		return nil, err
	}

	return comics, nil
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
		err := rows.Scan(&comic.ID, &comic.Page, &comic.Name, &comic.URL, &comic.ImageURL, &comic.LatestChap, &comic.ChapURL, &comic.Date, &comic.DateFormat)
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
