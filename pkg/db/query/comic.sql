-- name: CreateComic :one
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
	RETURNING *;

-- name: GetComic :one
SELECT * FROM comics
WHERE id = $1;

-- name: GetComicByURL :one
SELECT * FROM comics
WHERE url = $1;

-- name: GetComicByPageAndComicName :one
SELECT * FROM comics
WHERE comics.page=$1 AND comics.name=$2;

-- name: GetComicByPSIDAndComicID :one
SELECT comics.* FROM comics
JOIN subscribers ON comics.id=subscribers.comic_id
JOIN users ON users.id=subscribers.user_id
WHERE users.psid=$1 AND comics.id=$2;

-- name: GetComicForUpdate :one
SELECT * FROM comics
WHERE id = $1 FOR NO KEY UPDATE;

-- name: SearchComicOfUserByName :many
SELECT comics.* FROM comics
LEFT JOIN subscribers ON comics.id=subscribers.comic_id
WHERE subscribers.user_id=$1
AND (comics.name ILIKE $2 or unaccent(comics.name) ILIKE $2)
ORDER BY subscribers.created_at DESC;
-- LIMIT $3
-- OFFSET $4;

-- name: ListComics :many
SELECT * FROM comics
ORDER BY id DESC;
-- LIMIT $1
-- OFFSET $2;

-- name: ListComicsPerUser :many
SELECT comics.* FROM comics
LEFT JOIN subscribers ON comics.id=subscribers.comic_id 
WHERE subscribers.user_id=$1 ORDER BY subscribers.created_at DESC;
-- LIMIT $2
-- OFFSET $3;

-- name: UpdateComic :one
UPDATE comics 
SET latest_chap=$2, chap_url=$3, img_url=$4, cloud_img_url=$5 
WHERE id=$1
RETURNING *;

-- name: DeleteComic :exec
DELETE FROM comics
WHERE id = $1;
