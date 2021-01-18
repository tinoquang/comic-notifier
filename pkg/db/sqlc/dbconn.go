package db

import (
	"database/sql"

	_ "github.com/lib/pq" // don't use but still import for database/sql to init sql driver
	"github.com/tinoquang/comic-notifier/pkg/conf"
)

// NewDBConn return new DB connection
func NewDBConn() *sql.DB {

	Db, err := sql.Open("postgres", conf.Cfg.DBInfo)
	if err != nil {
		panic(err)
	}

	return Db
}
