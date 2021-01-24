// Code generated by sqlc. DO NOT EDIT.
// source: user.sql

package db

import (
	"context"
	"database/sql"
)

const createUser = `-- name: CreateUser :one
INSERT INTO users 
	(name,
	psid,
	appid,
	profile_pic) 
	VALUES ($1, $2, $3, $4)
	ON CONFLICT (psid) DO NOTHING
	RETURNING id, name, psid, appid, profile_pic
`

type CreateUserParams struct {
	Name       string
	Psid       sql.NullString
	Appid      sql.NullString
	ProfilePic sql.NullString
}

func (q *Queries) CreateUser(ctx context.Context, arg CreateUserParams) (User, error) {
	row := q.db.QueryRowContext(ctx, createUser,
		arg.Name,
		arg.Psid,
		arg.Appid,
		arg.ProfilePic,
	)
	var i User
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.Psid,
		&i.Appid,
		&i.ProfilePic,
	)
	return i, err
}

const deleteUser = `-- name: DeleteUser :exec
DELETE FROM users
WHERE psid = $1
`

func (q *Queries) DeleteUser(ctx context.Context, psid sql.NullString) error {
	_, err := q.db.ExecContext(ctx, deleteUser, psid)
	return err
}

const getUserByAppID = `-- name: GetUserByAppID :one
SELECT id, name, psid, appid, profile_pic FROM users
WHERE appid = $1
`

func (q *Queries) GetUserByAppID(ctx context.Context, appid sql.NullString) (User, error) {
	row := q.db.QueryRowContext(ctx, getUserByAppID, appid)
	var i User
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.Psid,
		&i.Appid,
		&i.ProfilePic,
	)
	return i, err
}

const getUserByPSID = `-- name: GetUserByPSID :one
SELECT id, name, psid, appid, profile_pic FROM users
WHERE psid = $1
`

func (q *Queries) GetUserByPSID(ctx context.Context, psid sql.NullString) (User, error) {
	row := q.db.QueryRowContext(ctx, getUserByPSID, psid)
	var i User
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.Psid,
		&i.Appid,
		&i.ProfilePic,
	)
	return i, err
}

const listUsers = `-- name: ListUsers :many
SELECT id, name, psid, appid, profile_pic FROM users
ORDER BY id
`

func (q *Queries) ListUsers(ctx context.Context) ([]User, error) {
	rows, err := q.db.QueryContext(ctx, listUsers)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []User{}
	for rows.Next() {
		var i User
		if err := rows.Scan(
			&i.ID,
			&i.Name,
			&i.Psid,
			&i.Appid,
			&i.ProfilePic,
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

const listUsersPerComic = `-- name: ListUsersPerComic :many
SELECT users.id, users.name, users.psid, users.appid, users.profile_pic FROM users
LEFT JOIN subscribers ON users.id=subscribers.user_id
WHERE subscribers.comic_id=$1 ORDER BY users.id DESC
`

func (q *Queries) ListUsersPerComic(ctx context.Context, comicID int32) ([]User, error) {
	rows, err := q.db.QueryContext(ctx, listUsersPerComic, comicID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []User{}
	for rows.Next() {
		var i User
		if err := rows.Scan(
			&i.ID,
			&i.Name,
			&i.Psid,
			&i.Appid,
			&i.ProfilePic,
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

const updateUser = `-- name: UpdateUser :one
UPDATE users
SET appid=$1
WHERE psid=$2
RETURNING id, name, psid, appid, profile_pic
`

type UpdateUserParams struct {
	Appid sql.NullString
	Psid  sql.NullString
}

func (q *Queries) UpdateUser(ctx context.Context, arg UpdateUserParams) (User, error) {
	row := q.db.QueryRowContext(ctx, updateUser, arg.Appid, arg.Psid)
	var i User
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.Psid,
		&i.Appid,
		&i.ProfilePic,
	)
	return i, err
}