-- name: CreateUser :one
INSERT INTO users 
	(name,
	psid,
	appid,
	profile_pic) 
	VALUES ($1, $2, $3, $4)
	ON CONFLICT (psid) DO NOTHING
	RETURNING *;


-- name: GetUserByPSID :one
SELECT * FROM users
WHERE psid = $1 LIMIT 1;

--name GetSubscribedComic :one
SELECT * FROM comics
LEFT JOIN subscribers ON comics.id=subscribers.comic_id 
WHERE subscribers.user_psid=$1 AND subscribers.comic_id=$2;

-- name: ListUsers :many
SELECT * FROM users
ORDER BY id;

-- name: DeleteUser :exec
DELETE FROM users
WHERE psid = $1;
