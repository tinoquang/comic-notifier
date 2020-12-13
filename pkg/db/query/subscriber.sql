-- name: CreateSubscriber :one
INSERT INTO subscribers
	(user_psid,
	comic_id) 
	VALUES ($1,$2)
	RETURNING *;

-- name: GetSubscriber :one
SELECT * FROM subscribers
WHERE user_psid=$1 AND comic_id=$2
LIMIT 1;

-- name: ListComicSubscribers :many
SELECT * FROM subscribers
WHERE comic_id=$1
ORDER BY id;

-- name: DeleteSubscriber :exec
DELETE FROM subscribers
WHERE user_psid=$1 AND comic_id=$2;
