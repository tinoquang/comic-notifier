package store

import (
	"context"
	"database/sql"
	"strings"

	"github.com/keegancsmith/sqlf"
	"github.com/tinoquang/comic-notifier/pkg/db"
	"github.com/tinoquang/comic-notifier/pkg/logging"
	"github.com/tinoquang/comic-notifier/pkg/model"
	"github.com/tinoquang/comic-notifier/pkg/util"
)

// ComicRepo contains comic's interact method
type ComicRepo interface {
	Get(ctx context.Context, id int) (*model.Comic, error)
	GetByURL(ctx context.Context, URL string) (*model.Comic, error)
	CheckComicSubscribe(ctx context.Context, psid string, comicID int) (*model.Comic, error)
	Create(ctx context.Context, comic *model.Comic) error
	Update(ctx context.Context, comic *model.Comic) error
	Delete(ctx context.Context, id int) error
	List(ctx context.Context, opt *ComicsListOptions) ([]model.Comic, error)
	ListByPSID(ctx context.Context, opt *ComicsListOptions, psid string) ([]model.Comic, error)
}

type comicDB struct {
	dbconn *sql.DB
}

func newComicRepo(dbconn *sql.DB) *comicDB {
	return &comicDB{dbconn: dbconn}
}

// ComicsListOptions specifies the options for listing projects.
type ComicsListOptions struct {
	*NameLikeOptions
	*LimitOffset
}

// LimitOffset specifies SQL LIMIT and OFFSET counts. A pointer to it is typically embedded in other options
// structures that need to perform SQL queries with LIMIT and OFFSET.
type LimitOffset struct {
	Limit  int // SQL LIMIT count
	Offset int // SQL OFFSET count
}

// SQL returns the SQL query fragment ("LIMIT %d OFFSET %d") for use in SQL queries.
func (o *LimitOffset) SQL() *sqlf.Query {
	if o == nil {
		return &sqlf.Query{}
	}

	if o.Limit == 0 {
		return sqlf.Sprintf("LIMIT ALL OFFSET %d", o.Offset)
	}

	return sqlf.Sprintf("LIMIT %d OFFSET %d", o.Limit, o.Offset)
}

// NameLikeOptions used to query by name using like
type NameLikeOptions struct {
	// Query specifies a search query for organizations.
	Query string
}

// ListComicNameLikeSQL used to search by name if query is set
func ListComicNameLikeSQL(opt *NameLikeOptions) (conds []*sqlf.Query) {
	conds = []*sqlf.Query{sqlf.Sprintf("TRUE")}
	if opt.Query != "" {
		query := "%" + strings.Replace(opt.Query, " ", "%", -1) + "%"
		conds = append(conds, sqlf.Sprintf("comics.name ILIKE %s or unaccent(comics.name) ILIKE %s", query, query))
	}
	return conds
}

// NewComicsListOptions create a new opts
func NewComicsListOptions(query string, limit int, offset int) *ComicsListOptions {
	return &ComicsListOptions{
		NameLikeOptions: &NameLikeOptions{query},
		LimitOffset:     &LimitOffset{Limit: limit, Offset: offset},
	}
}

func (c *comicDB) Get(ctx context.Context, id int) (*model.Comic, error) {

	comics, err := c.getBySQL(ctx, "WHERE id=$1", id)
	if err != nil {
		logging.Danger()
		return nil, err
	}

	if len(comics) == 0 {
		return &model.Comic{}, util.ErrNotFound
	}

	return &comics[0], nil
}

func (c *comicDB) GetByURL(ctx context.Context, URL string) (*model.Comic, error) {

	comics, err := c.getBySQL(ctx, "WHERE url=$1", URL)
	if err != nil {
		logging.Danger()
		return nil, err
	}

	if len(comics) == 0 {
		return &model.Comic{}, util.ErrNotFound
	}

	return &comics[0], nil
}

// CheckComicSubscribe check if user has already subscribed to comic
func (c *comicDB) CheckComicSubscribe(ctx context.Context, psid string, comicID int) (*model.Comic, error) {

	query := "LEFT JOIN subscribers ON comics.id=subscribers.comic_id WHERE subscribers.user_psid=$1 AND subscribers.comic_id=$2"

	comics, err := c.getBySQL(ctx, query, psid, comicID)
	if err != nil {
		logging.Danger(err)
		return nil, err
	}

	if len(comics) == 0 {
		return &model.Comic{}, util.ErrNotFound
	}

	return &comics[0], nil
}

func (c *comicDB) Create(ctx context.Context, comic *model.Comic) error {

	query := "INSERT INTO comics (page, name, url, img_url, cloud_img, latest_chap, chap_url) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id"

	err := db.WithTransaction(ctx, c.dbconn, func(tx db.Transaction) error {
		return tx.QueryRowContext(
			ctx, query, comic.Page, comic.Name, comic.URL, comic.OriginImgURL, comic.CloudImg, comic.LatestChap, comic.ChapURL,
		).Scan(&comic.ID)
	})
	return err

}

func (c *comicDB) Update(ctx context.Context, comic *model.Comic) error {

	query := "UPDATE comics SET latest_chap=$2, chap_url=$3, img_url=$4, cloud_img=$5 WHERE id=$1"
	_, err := c.dbconn.ExecContext(ctx, query, comic.ID, comic.LatestChap, comic.ChapURL, comic.OriginImgURL, comic.CloudImg)
	return err
}

func (c *comicDB) Delete(ctx context.Context, id int) error {

	query := "DELETE FROM comics WHERE id=$1"
	_, err := c.dbconn.ExecContext(ctx, query, id)
	if err != nil {
		logging.Danger(err)
	}
	return err
}

func (c *comicDB) List(ctx context.Context, opt *ComicsListOptions) ([]model.Comic, error) {

	if opt == nil {
		opt = &ComicsListOptions{}
	}

	conds := ListComicNameLikeSQL(opt.NameLikeOptions)

	q := sqlf.Sprintf("WHERE %s ORDER BY id DESC %s", sqlf.Join(conds, "AND"), opt.LimitOffset.SQL())

	return c.getBySQL(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...)
}

// ListByPSID used to list all comics of specific user
func (c *comicDB) ListByPSID(ctx context.Context, opt *ComicsListOptions, psid string) ([]model.Comic, error) {

	if opt == nil {
		opt = &ComicsListOptions{}
	}

	conds := ListComicNameLikeSQL(opt.NameLikeOptions)
	conds = append(conds, sqlf.Sprintf("subscribers.user_psid = %s", psid))

	q := sqlf.Sprintf("LEFT JOIN subscribers ON comics.id=subscribers.comic_id WHERE %s ORDER BY subscribers.created_at DESC %s", sqlf.Join(conds, "AND"), opt.LimitOffset.SQL())

	comics, err := c.getBySQL(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...)
	if err != nil {
		logging.Danger(err)
		return nil, err
	}

	if len(comics) == 0 {
		return []model.Comic{}, util.ErrNotFound
	}

	return comics, nil
}

func (c *comicDB) getBySQL(ctx context.Context, query string, args ...interface{}) ([]model.Comic, error) {
	rows, err := c.dbconn.QueryContext(ctx, "SELECT comics.* FROM comics "+query, args...)
	if err != nil {
		logging.Danger()
		return nil, err
	}

	comics := []model.Comic{}
	defer rows.Close()
	for rows.Next() {
		comic := model.Comic{}
		err := rows.Scan(&comic.ID, &comic.Page, &comic.Name, &comic.URL, &comic.OriginImgURL, &comic.CloudImg, &comic.LatestChap, &comic.ChapURL)
		if err != nil {
			logging.Danger(err)
			return nil, err
		}

		comics = append(comics, comic)
	}
	if err = rows.Err(); err != nil {
		logging.Danger()
		return nil, err
	}

	return comics, nil
}
