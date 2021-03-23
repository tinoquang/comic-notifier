// Code generated by sqlc. DO NOT EDIT.
// source: comic.sql

package db

import (
	"context"
	"database/sql"
)

const createComic = `-- name: CreateComic :one
INSERT INTO comics
	(page,
	name,
	url,
	img_url,
	cloud_img_url,
	latest_chap,
	chap_url)
	VALUES ($1,$2,$3,$4,$5,$6,$7)
	ON CONFLICT (url) DO NOTHING
	RETURNING id, page, name, url, img_url, cloud_img_url, latest_chap, chap_url
`

type CreateComicParams struct {
	Page        string
	Name        string
	Url         string
	ImgUrl      string
	CloudImgUrl string
	LatestChap  string
	ChapUrl     string
}

func (q *Queries) CreateComic(ctx context.Context, arg CreateComicParams) (Comic, error) {
	row := q.db.QueryRowContext(ctx, createComic,
		arg.Page,
		arg.Name,
		arg.Url,
		arg.ImgUrl,
		arg.CloudImgUrl,
		arg.LatestChap,
		arg.ChapUrl,
	)
	var i Comic
	err := row.Scan(
		&i.ID,
		&i.Page,
		&i.Name,
		&i.Url,
		&i.ImgUrl,
		&i.CloudImgUrl,
		&i.LatestChap,
		&i.ChapUrl,
	)
	return i, err
}

const deleteComic = `-- name: DeleteComic :exec
DELETE FROM comics
WHERE id = $1
`

func (q *Queries) DeleteComic(ctx context.Context, id int32) error {
	_, err := q.db.ExecContext(ctx, deleteComic, id)
	return err
}

const getComic = `-- name: GetComic :one
SELECT id, page, name, url, img_url, cloud_img_url, latest_chap, chap_url FROM comics
WHERE id = $1
`

func (q *Queries) GetComic(ctx context.Context, id int32) (Comic, error) {
	row := q.db.QueryRowContext(ctx, getComic, id)
	var i Comic
	err := row.Scan(
		&i.ID,
		&i.Page,
		&i.Name,
		&i.Url,
		&i.ImgUrl,
		&i.CloudImgUrl,
		&i.LatestChap,
		&i.ChapUrl,
	)
	return i, err
}

const getComicByPSIDAndComicID = `-- name: GetComicByPSIDAndComicID :one
SELECT comics.id, comics.page, comics.name, comics.url, comics.img_url, comics.cloud_img_url, comics.latest_chap, comics.chap_url FROM comics
JOIN subscribers ON comics.id=subscribers.comic_id
JOIN users ON users.id=subscribers.user_id
WHERE users.psid=$1 AND comics.id=$2
`

type GetComicByPSIDAndComicIDParams struct {
	Psid sql.NullString
	ID   int32
}

func (q *Queries) GetComicByPSIDAndComicID(ctx context.Context, arg GetComicByPSIDAndComicIDParams) (Comic, error) {
	row := q.db.QueryRowContext(ctx, getComicByPSIDAndComicID, arg.Psid, arg.ID)
	var i Comic
	err := row.Scan(
		&i.ID,
		&i.Page,
		&i.Name,
		&i.Url,
		&i.ImgUrl,
		&i.CloudImgUrl,
		&i.LatestChap,
		&i.ChapUrl,
	)
	return i, err
}

const getComicByPageAndComicName = `-- name: GetComicByPageAndComicName :one
SELECT id, page, name, url, img_url, cloud_img_url, latest_chap, chap_url FROM comics
WHERE comics.page=$1 AND comics.name=$2
`

type GetComicByPageAndComicNameParams struct {
	Page string
	Name string
}

func (q *Queries) GetComicByPageAndComicName(ctx context.Context, arg GetComicByPageAndComicNameParams) (Comic, error) {
	row := q.db.QueryRowContext(ctx, getComicByPageAndComicName, arg.Page, arg.Name)
	var i Comic
	err := row.Scan(
		&i.ID,
		&i.Page,
		&i.Name,
		&i.Url,
		&i.ImgUrl,
		&i.CloudImgUrl,
		&i.LatestChap,
		&i.ChapUrl,
	)
	return i, err
}

const getComicByURL = `-- name: GetComicByURL :one
SELECT id, page, name, url, img_url, cloud_img_url, latest_chap, chap_url FROM comics
WHERE url = $1
`

func (q *Queries) GetComicByURL(ctx context.Context, url string) (Comic, error) {
	row := q.db.QueryRowContext(ctx, getComicByURL, url)
	var i Comic
	err := row.Scan(
		&i.ID,
		&i.Page,
		&i.Name,
		&i.Url,
		&i.ImgUrl,
		&i.CloudImgUrl,
		&i.LatestChap,
		&i.ChapUrl,
	)
	return i, err
}

const getComicForUpdate = `-- name: GetComicForUpdate :one
SELECT id, page, name, url, img_url, cloud_img_url, latest_chap, chap_url FROM comics
WHERE id = $1 FOR NO KEY UPDATE
`

func (q *Queries) GetComicForUpdate(ctx context.Context, id int32) (Comic, error) {
	row := q.db.QueryRowContext(ctx, getComicForUpdate, id)
	var i Comic
	err := row.Scan(
		&i.ID,
		&i.Page,
		&i.Name,
		&i.Url,
		&i.ImgUrl,
		&i.CloudImgUrl,
		&i.LatestChap,
		&i.ChapUrl,
	)
	return i, err
}

const listComics = `-- name: ListComics :many

SELECT id, page, name, url, img_url, cloud_img_url, latest_chap, chap_url FROM comics
ORDER BY id DESC
`

// LIMIT $3
// OFFSET $4;
func (q *Queries) ListComics(ctx context.Context) ([]Comic, error) {
	rows, err := q.db.QueryContext(ctx, listComics)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []Comic{}
	for rows.Next() {
		var i Comic
		if err := rows.Scan(
			&i.ID,
			&i.Page,
			&i.Name,
			&i.Url,
			&i.ImgUrl,
			&i.CloudImgUrl,
			&i.LatestChap,
			&i.ChapUrl,
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

const listComicsPerUserPSID = `-- name: ListComicsPerUserPSID :many

SELECT comics.id, comics.page, comics.name, comics.url, comics.img_url, comics.cloud_img_url, comics.latest_chap, comics.chap_url FROM comics
LEFT JOIN subscribers ON comics.id=subscribers.comic_id 
WHERE subscribers.user_id=$1 ORDER BY subscribers.created_at DESC
`

// LIMIT $1
// OFFSET $2;
func (q *Queries) ListComicsPerUserPSID(ctx context.Context, userID int32) ([]Comic, error) {
	rows, err := q.db.QueryContext(ctx, listComicsPerUserPSID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []Comic{}
	for rows.Next() {
		var i Comic
		if err := rows.Scan(
			&i.ID,
			&i.Page,
			&i.Name,
			&i.Url,
			&i.ImgUrl,
			&i.CloudImgUrl,
			&i.LatestChap,
			&i.ChapUrl,
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

const searchComicOfUserByName = `-- name: SearchComicOfUserByName :many
SELECT comics.id, comics.page, comics.name, comics.url, comics.img_url, comics.cloud_img_url, comics.latest_chap, comics.chap_url FROM comics
LEFT JOIN subscribers ON comics.id=subscribers.comic_id
WHERE subscribers.user_id=$1
AND (comics.name ILIKE $2 or unaccent(comics.name) ILIKE $2)
ORDER BY subscribers.created_at DESC
`

type SearchComicOfUserByNameParams struct {
	UserID int32
	Name   string
}

func (q *Queries) SearchComicOfUserByName(ctx context.Context, arg SearchComicOfUserByNameParams) ([]Comic, error) {
	rows, err := q.db.QueryContext(ctx, searchComicOfUserByName, arg.UserID, arg.Name)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []Comic{}
	for rows.Next() {
		var i Comic
		if err := rows.Scan(
			&i.ID,
			&i.Page,
			&i.Name,
			&i.Url,
			&i.ImgUrl,
			&i.CloudImgUrl,
			&i.LatestChap,
			&i.ChapUrl,
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

const updateComic = `-- name: UpdateComic :one

UPDATE comics 
SET latest_chap=$2, chap_url=$3, img_url=$4, cloud_img_url=$5 
WHERE id=$1
RETURNING id, page, name, url, img_url, cloud_img_url, latest_chap, chap_url
`

type UpdateComicParams struct {
	ID          int32
	LatestChap  string
	ChapUrl     string
	ImgUrl      string
	CloudImgUrl string
}

// LIMIT $2
// OFFSET $3;
func (q *Queries) UpdateComic(ctx context.Context, arg UpdateComicParams) (Comic, error) {
	row := q.db.QueryRowContext(ctx, updateComic,
		arg.ID,
		arg.LatestChap,
		arg.ChapUrl,
		arg.ImgUrl,
		arg.CloudImgUrl,
	)
	var i Comic
	err := row.Scan(
		&i.ID,
		&i.Page,
		&i.Name,
		&i.Url,
		&i.ImgUrl,
		&i.CloudImgUrl,
		&i.LatestChap,
		&i.ChapUrl,
	)
	return i, err
}
