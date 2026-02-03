-- name: CreatePost :one
INSERT INTO posts (id, created_at, updated_at, title, url, description, published_at, feed_id)
VALUES (gen_random_uuid(), now(), now(), $1, $2, $3, $4, $5)
RETURNING *;

-- name: GetPostFromUser :many
SELECT p.title, p.url, p.description, p.published_at FROM posts p JOIN feed_follows ff ON p.feed_id = ff.feed_id WHERE ff.user_id=$1 ORDER BY p.published_at DESC LIMIT $2;
