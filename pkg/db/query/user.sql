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
WHERE psid = $1;

-- name: GetUserByAppID :one
SELECT * FROM users
WHERE appid = $1;

-- name: ListUsers :many
SELECT * FROM users
ORDER BY id;

-- name: ListUsersPerComic :many
SELECT users.* FROM users
LEFT JOIN subscribers ON users.id=subscribers.user_id
WHERE subscribers.comic_id=$1 ORDER BY users.id DESC;

-- name: UpdateUser :one
UPDATE users
SET appid=$1
WHERE psid=$2
RETURNING *;

-- name: DeleteUser :exec
DELETE FROM users
WHERE psid = $1;
