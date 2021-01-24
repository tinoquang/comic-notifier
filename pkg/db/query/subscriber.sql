-- name: CreateSubscriber :one
INSERT INTO subscribers
	(user_id,
	comic_id) 
	VALUES ($1,$2)
	RETURNING *;

-- name: GetSubscriber :one
SELECT * FROM subscribers
WHERE user_id=$1 AND comic_id=$2;

-- name: DeleteSubscriber :exec
DELETE FROM subscribers
WHERE user_id=$1 AND comic_id=$2;
