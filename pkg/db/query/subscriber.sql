-- name: CreateSubscriber :one
INSERT INTO subscribers
	(user_psid,
	user_appid,
	comic_id) 
	VALUES ($1,$2,$3)
	RETURNING *;

-- name: GetSubscriberByPSIDAndComicID :one
SELECT * FROM subscribers
WHERE user_psid=$1 AND comic_id=$2;

-- name: GetSubscriberByAppIDAndComicID :one
SELECT * FROM subscribers
WHERE user_appid=$1 AND comic_id=$2;

-- name: ListSubscriberByComicID :many
SELECT * FROM subscribers
WHERE subscribers.comic_id=$1;

-- name: DeleteSubscriberByAppID :exec
DELETE FROM subscribers
WHERE user_appid=$1 AND comic_id=$2;

-- name: DeleteSubscriberByPSID :exec
DELETE FROM subscribers
WHERE user_psid=$1 AND comic_id=$2;
