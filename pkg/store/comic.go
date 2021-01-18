package store

import (
	"strings"

	"github.com/keegancsmith/sqlf"
)

/*-------------------------- Handle query options ------------------------------- */

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
