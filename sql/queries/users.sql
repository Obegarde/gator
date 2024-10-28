
-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, name)
VALUES (
    $1,
    $2,
    $3,
    $4
)
RETURNING *;

-- name: CreateFeed :one
INSERT INTO feeds (id, created_at,updated_at,user_id,name,url)
VALUES (
	$1,
	$2,
	$3,
	$4,
	$5,
	$6
	)
RETURNING *;


-- name: GetUser :one
SELECT * FROM users
WHERE name = $1;

-- name: ResetUsers :exec
DELETE FROM users;

-- name: GetUsers :many
SELECT * FROM users;

-- name: GetFeeds :many
SELECT * FROM feeds;

-- name: GetUserByID :one
SELECT * FROM users
WHERE id = $1;
