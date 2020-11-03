package store

import (
	"context"
	"database/sql"

	"github.com/pkg/errors"
	"github.com/tinoquang/comic-notifier/pkg/conf"
	"github.com/tinoquang/comic-notifier/pkg/db"
	"github.com/tinoquang/comic-notifier/pkg/logging"
	"github.com/tinoquang/comic-notifier/pkg/model"
)

// SubscriberInterface contains subscriber's interact method
type SubscriberInterface interface {
	Get(ctx context.Context, psid string, comicID int) (*model.Subscriber, error)
	Create(ctx context.Context, subscriber *model.Subscriber) error
	Delete(ctx context.Context, psid string, comicID int) error
	ListByComicID(ctx context.Context, comicID int) ([]model.Subscriber, error)
}

type subscriberDB struct {
	dbconn *sql.DB
	cfg    *conf.Config
}

// NewSubscriberStore return subscriber interfaces
func NewSubscriberStore(dbconn *sql.DB, cfg *conf.Config) SubscriberInterface {
	return &subscriberDB{dbconn: dbconn, cfg: cfg}
}

func (s *subscriberDB) Get(ctx context.Context, psid string, comicID int) (*model.Subscriber, error) {

	subscribers, err := s.getBySQL(ctx, "WHERE user_psid=$1 AND comic_id=$2", psid, comicID)
	if err != nil {
		logging.Danger()
		return nil, err
	}

	if len(subscribers) == 0 {
		return &model.Subscriber{}, errors.New("Subscriber not found")
	}

	return &subscribers[0], nil
}

func (s *subscriberDB) Create(ctx context.Context, subscriber *model.Subscriber) error {

	query := "INSERT INTO subscribers (user_psid, comic_id) VALUES ($1,$2) RETURNING id"

	err := db.WithTransaction(ctx, s.dbconn, func(tx db.Transaction) error {
		return tx.QueryRowContext(
			ctx, query, subscriber.PSID, subscriber.ComicID,
		).Scan(&subscriber.ID)
	})
	return err
}

func (s *subscriberDB) ListByComicID(ctx context.Context, comicID int) ([]model.Subscriber, error) {

	return s.getBySQL(ctx, "WHERE comic_id=$1", comicID)
}

func (s *subscriberDB) Delete(ctx context.Context, psid string, comicID int) error {

	query := "DELETE FROM subscribers WHERE user_psid=$1 AND comic_id=$2"
	_, err := s.dbconn.ExecContext(ctx, query, psid, comicID)
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
		err := rows.Scan(&subscriber.ID, &subscriber.PSID, &subscriber.ComicID)
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
