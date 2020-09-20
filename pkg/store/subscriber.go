package store

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/pkg/errors"
	"github.com/tinoquang/comic-notifier/pkg/conf"
	"github.com/tinoquang/comic-notifier/pkg/model"
)

// SubscriberInterface contains subscriber's interact method
type SubscriberInterface interface {
	Get(ctx context.Context, field string, id string) (*model.Subscriber, error)
}

type subscriberDB struct {
	dbconn *sql.DB
	cfg    *conf.Config
}

// NewSubscriberStore return subscriber interfaces
func NewSubscriberStore(dbconn *sql.DB, cfg *conf.Config) SubscriberInterface {
	return &subscriberDB{dbconn: dbconn, cfg: cfg}
}

func (s *subscriberDB) Get(ctx context.Context, field string, id string) (*model.Subscriber, error) {
	query := "WHERE " + field + "=$1 LIMIT 1"
	subscribers, err := s.getBySQL(ctx, query, id)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get customer by id: %s", id)
	}

	if len(subscribers) == 0 {
		return nil, errors.New(fmt.Sprintf("subscriber %s not found", id))
	}

	return &subscribers[0], nil
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
		err := rows.Scan(&subscriber.ID, &subscriber.UserID, &subscriber.ComicID)
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
