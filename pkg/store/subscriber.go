package store

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/pkg/errors"
	"github.com/tinoquang/comic-notifier/pkg/conf"
	"github.com/tinoquang/comic-notifier/pkg/db"
	"github.com/tinoquang/comic-notifier/pkg/model"
)

// SubscriberInterface contains subscriber's interact method
type SubscriberInterface interface {
	Get(ctx context.Context, userid, comicid int) (*model.Subscriber, error)
	GetByID(ctx context.Context, id int) (*model.Subscriber, error)
	Create(ctx context.Context, subscriber *model.Subscriber) error
	Delete(ctx context.Context, id int) error
}

type subscriberDB struct {
	dbconn *sql.DB
	cfg    *conf.Config
}

// NewSubscriberStore return subscriber interfaces
func NewSubscriberStore(dbconn *sql.DB, cfg *conf.Config) SubscriberInterface {
	return &subscriberDB{dbconn: dbconn, cfg: cfg}
}

func (s *subscriberDB) Get(ctx context.Context, userid, comicid int) (*model.Subscriber, error) {
	query := "WHERE user_id=$1 AND comic_id=$2"

	subscribers, err := s.getBySQL(ctx, query, userid, comicid)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get subscriber")
	}

	if len(subscribers) == 0 {
		return &model.Subscriber{}, errors.New(fmt.Sprintf("subscriber with userid = %d and comicid = %d not found", userid, comicid))
	}

	return &subscribers[0], nil
}

func (s *subscriberDB) GetByID(ctx context.Context, id int) (*model.Subscriber, error) {

	query := "WHERE id=$1"

	subscribers, err := s.getBySQL(ctx, query, id)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get subscriber")
	}

	if len(subscribers) == 0 {
		return &model.Subscriber{}, errors.New(fmt.Sprintf("subscriber with id = %d not found", id))
	}

	return &subscribers[0], nil
}

func (s *subscriberDB) Create(ctx context.Context, subscriber *model.Subscriber) error {

	query := "INSERT INTO subscribers (page, user_id, username, comic_id, comicname) VALUES ($1,$2,$3,$4,$5) RETURNING id"

	err := db.WithTransaction(ctx, s.dbconn, func(tx db.Transaction) error {
		return tx.QueryRowContext(
			ctx, query, subscriber.Page, subscriber.UserID, subscriber.UserName, subscriber.ComicID, subscriber.ComicName,
		).Scan(&subscriber.ID)
	})
	return err
}

func (s *subscriberDB) Delete(ctx context.Context, id int) error {

	query := "DELETE FROM subscribers WHERE id=$1"
	_, err := s.dbconn.ExecContext(ctx, query, id)
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
		err := rows.Scan(&subscriber.ID, &subscriber.Page, &subscriber.UserID, &subscriber.UserName, &subscriber.ComicID, &subscriber.ComicName)
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
