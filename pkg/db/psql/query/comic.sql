-- name: CreateComic :one
INSERT INTO comics
	(page,
	name,
	url,
	img_url,
	cloud_img,
	latest_chap,
	chap_url)
	VALUES ($1,$2,$3,$4,$5,$6,$7)
	RETURNING *;

-- name: GetComic :one
SELECT * FROM comics
WHERE id = $1 LIMIT 1;

-- name: GetComicByURL :one
SELECT * FROM comics
WHERE url = $1 LIMIT 1;

--name GetSubscribedComic :one
SELECT * FROM comics
LEFT JOIN subscribers ON comics.id=subscribers.comic_id 
WHERE subscribers.user_psid=$1 AND subscribers.comic_id=$2;

-- name: GetComicForUpdate :one
SELECT * FROM comics
WHERE id = $1LIMIT 1
FOR NO KEY UPDATE;

-- name: ListComics :many
SELECT * FROM comics
WHERE comics.name ILIKE $1 or unaccent(comics.name) ILIKE $2
ORDER BY id DESC
LIMIT $3
OFFSET $4;

--name: ListComicsByUserPSID :many
SELECT * FROM comics
LEFT JOIN subscribers ON comics.id=subscribers.comic_id 
WHERE subscribers.user_psid=$1 ORDER BY subscribers.created_at DESC
LIMIT $2
OFFSET $3;

-- name: UpdateComic :one
UPDATE comics 
SET latest_chap=$2, chap_url=$3, img_url=$4, cloud_img=$5 
WHERE id=$1
RETURNING *;

-- name: DeleteComic :exec
DELETE FROM comics
WHERE id = $1;
