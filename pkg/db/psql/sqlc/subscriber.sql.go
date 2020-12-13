// Code generated by sqlc. DO NOT EDIT.
// source: subscriber.sql

package db

import (
	"context"
	"database/sql"
)

const createSubscriber = `-- name: CreateSubscriber :one
INSERT INTO subscribers
	(user_psid,
	comic_id) 
	VALUES ($1,$2)
	RETURNING id, user_psid, comic_id, created_at
`

type CreateSubscriberParams struct {
	UserPsid sql.NullString
	ComicID  sql.NullInt32
}

func (q *Queries) CreateSubscriber(ctx context.Context, arg CreateSubscriberParams) (Subscriber, error) {
	row := q.db.QueryRowContext(ctx, createSubscriber, arg.UserPsid, arg.ComicID)
	var i Subscriber
	err := row.Scan(
		&i.ID,
		&i.UserPsid,
		&i.ComicID,
		&i.CreatedAt,
	)
	return i, err
}

const deleteSubscriber = `-- name: DeleteSubscriber :exec
DELETE FROM subscribers
WHERE user_psid=$1 AND comic_id=$2
`

type DeleteSubscriberParams struct {
	UserPsid sql.NullString
	ComicID  sql.NullInt32
}

func (q *Queries) DeleteSubscriber(ctx context.Context, arg DeleteSubscriberParams) error {
	_, err := q.db.ExecContext(ctx, deleteSubscriber, arg.UserPsid, arg.ComicID)
	return err
}

const getSubscriber = `-- name: GetSubscriber :one
SELECT id, user_psid, comic_id, created_at FROM subscribers
WHERE user_psid=$1 AND comic_id=$2
LIMIT 1
`

type GetSubscriberParams struct {
	UserPsid sql.NullString
	ComicID  sql.NullInt32
}

func (q *Queries) GetSubscriber(ctx context.Context, arg GetSubscriberParams) (Subscriber, error) {
	row := q.db.QueryRowContext(ctx, getSubscriber, arg.UserPsid, arg.ComicID)
	var i Subscriber
	err := row.Scan(
		&i.ID,
		&i.UserPsid,
		&i.ComicID,
		&i.CreatedAt,
	)
	return i, err
}

const listComicSubscribers = `-- name: ListComicSubscribers :many
SELECT id, user_psid, comic_id, created_at FROM subscribers
WHERE comic_id=$1
ORDER BY id
`

func (q *Queries) ListComicSubscribers(ctx context.Context, comicID sql.NullInt32) ([]Subscriber, error) {
	rows, err := q.db.QueryContext(ctx, listComicSubscribers, comicID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []Subscriber{}
	for rows.Next() {
		var i Subscriber
		if err := rows.Scan(
			&i.ID,
			&i.UserPsid,
			&i.ComicID,
			&i.CreatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}
