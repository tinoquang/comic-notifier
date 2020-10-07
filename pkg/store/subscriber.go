package store

import (
	"context"
	"database/sql"

	"github.com/tinoquang/comic-notifier/pkg/conf"
	"github.com/tinoquang/comic-notifier/pkg/db"
	"github.com/tinoquang/comic-notifier/pkg/model"
)

// SubscriberInterface contains subscriber's interact method
type SubscriberInterface interface {
	Create(ctx context.Context, subscriber *model.Subscriber) error
	Delete(ctx context.Context, psid string, comicid int) error
}

type subscriberDB struct {
	dbconn *sql.DB
	cfg    *conf.Config
}

// NewSubscriberStore return subscriber interfaces
func NewSubscriberStore(dbconn *sql.DB, cfg *conf.Config) SubscriberInterface {
	return &subscriberDB{dbconn: dbconn, cfg: cfg}
}

func (s *subscriberDB) Create(ctx context.Context, subscriber *model.Subscriber) error {

	query := "INSERT INTO subscribers (user_psid, comic_id) VALUES ($1,$2)"

	err := db.WithTransaction(ctx, s.dbconn, func(tx db.Transaction) error {
		return tx.QueryRowContext(
			ctx, query, subscriber.PSID, subscriber.ComicID,
		).Scan(&subscriber.ID)
	})
	return err
}

func (s *subscriberDB) Delete(ctx context.Context, psid string, comicid int) error {

	query := "DELETE FROM subscribers WHERE user_psid=$1 AND comic_id=$2"
	_, err := s.dbconn.ExecContext(ctx, query, psid, comicid)
	return err
}

func (s *subscriberDB) getBySQL(ctx context.Context, query string, args ...interface{}) ([]model.Subscriber, error) {
	rows, err := s.dbconn.QueryContext(ctx, "SELECT * FROM subscribers "+query, args...)
	if err != nil {
		return nil, err
	}

	subscribers := []model.Subscriber{}
	defer rows.Close()
	for rows.Next() {
		subscriber := model.Subscriber{}
		err := rows.Scan(&subscriber.PSID, &subscriber.ComicID)
		if err != nil {
			return nil, err
		}

		subscribers = append(subscribers, subscriber)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return subscribers, nil
}
