package store

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/pkg/errors"
	"github.com/tinoquang/comic-notifier/pkg/conf"
	"github.com/tinoquang/comic-notifier/pkg/model"
)

// PageInterface contain page's interact method
type PageInterface interface {
	GetByName(ctx context.Context, name string) (*model.Page, error)
}

type pageDB struct {
	dbconn *sql.DB
	cfg    *conf.Config
}

// NewPageStore return page interfaces
func NewPageStore(dbconn *sql.DB, cfg *conf.Config) PageInterface {
	p := &pageDB{dbconn: dbconn, cfg: cfg}
	p.init()
	return p
}

func (p *pageDB) init() {
	_, err := p.dbconn.Exec("delete from pages")
	if err != nil {
		panic(fmt.Sprintf("Can't delete pages table: %s", err))
	}

	_, err = p.dbconn.Exec("alter sequence pages_id_seq restart with 1")
	if err != nil {
		panic(fmt.Sprintf("Can't reset pages id to 1: %s", err))
	}

	stmt, err := p.dbconn.Prepare("insert into pages (name) values ($1)")

	if err != nil {
		panic(fmt.Sprintf("Can't page table insert statement: %s", err))
	}

	defer stmt.Close()

	for _, page := range p.cfg.PageSupport.Page {
		_, err = stmt.Exec(page.Name)
		if err != nil {
			panic(err)
		}
	}

	return
}

// GetByName get page info by name
func (p *pageDB) GetByName(ctx context.Context, name string) (*model.Page, error) {

	pages, err := p.getBySQL(ctx, "WHERE name=$1 LIMIT 1", name)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get page with name: %s\n", name)
	}

	if len(pages) == 0 {
		return nil, errors.New(fmt.Sprintf("Page %s not found", name))
	}

	return &pages[0], nil
}

// func (p *pageDB) List(ctx context.Context) ([]model.Page, error) {
//
// }

func (p *pageDB) getBySQL(ctx context.Context, query string, args ...interface{}) ([]model.Page, error) {
	rows, err := p.dbconn.QueryContext(ctx, "SELECT * FROM pages "+query, args...)
	if err != nil {
		return nil, err
	}

	pages := []model.Page{}
	defer rows.Close()
	for rows.Next() {
		page := model.Page{}
		err := rows.Scan(&page.ID, &page.Name)
		if err != nil {
			return nil, err
		}

		pages = append(pages, page)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return pages, nil
}
