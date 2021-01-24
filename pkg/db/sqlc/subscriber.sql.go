// Code generated by sqlc. DO NOT EDIT.
// source: subscriber.sql

package db

import (
	"context"
)

const createSubscriber = `-- name: CreateSubscriber :one
INSERT INTO subscribers
	(user_id,
	comic_id) 
	VALUES ($1,$2)
	RETURNING id, user_id, comic_id, created_at
`

type CreateSubscriberParams struct {
	UserID  int32
	ComicID int32
}

func (q *Queries) CreateSubscriber(ctx context.Context, arg CreateSubscriberParams) (Subscriber, error) {
	row := q.db.QueryRowContext(ctx, createSubscriber, arg.UserID, arg.ComicID)
	var i Subscriber
	err := row.Scan(
		&i.ID,
		&i.UserID,
		&i.ComicID,
		&i.CreatedAt,
	)
	return i, err
}

const deleteSubscriber = `-- name: DeleteSubscriber :exec
DELETE FROM subscribers
WHERE user_id=$1 AND comic_id=$2
`

type DeleteSubscriberParams struct {
	UserID  int32
	ComicID int32
}

func (q *Queries) DeleteSubscriber(ctx context.Context, arg DeleteSubscriberParams) error {
	_, err := q.db.ExecContext(ctx, deleteSubscriber, arg.UserID, arg.ComicID)
	return err
}

const getSubscriber = `-- name: GetSubscriber :one
SELECT id, user_id, comic_id, created_at FROM subscribers
WHERE user_id=$1 AND comic_id=$2
`

type GetSubscriberParams struct {
	UserID  int32
	ComicID int32
}

func (q *Queries) GetSubscriber(ctx context.Context, arg GetSubscriberParams) (Subscriber, error) {
	row := q.db.QueryRowContext(ctx, getSubscriber, arg.UserID, arg.ComicID)
	var i Subscriber
	err := row.Scan(
		&i.ID,
		&i.UserID,
		&i.ComicID,
		&i.CreatedAt,
	)
	return i, err
}
