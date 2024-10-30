
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

-- name: CreateFeedFollow :many
WITH inserted_feed_follow AS(
INSERT INTO feed_follows (id, created_at,updated_at,user_id,feed_id)
VALUES(
	$1,
	$2,
	$3,
	$4,
	$5
	)
RETURNING *
)
SELECT
	inserted_feed_follow.*,
	feeds.name AS feed_name,
	users.name AS user_name
FROM inserted_feed_follow
INNER JOIN users
ON inserted_feed_follow.user_id = users.id
INNER JOIN feeds
ON inserted_feed_follow.feed_id = feeds.id;

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

-- name: GetFeedByURL :one
SELECT * FROM feeds 
WHERE url = $1;

-- name: GetFeedFollowsForUser :many
SELECT feed_follows.*,users.name AS user_name ,feeds.name AS feed_name 
FROM feed_follows
INNER JOIN feeds ON feed_follows.feed_id = feeds.id
INNER JOIN users ON feed_follows.user_id = users.id
WHERE	users.name = $1;



