-- name: AddFeed :one
INSERT INTO feeds(id, created_at, updated_at, name, url, user_id)
VALUES (gen_random_uuid(), now(), now(), $1, $2, $3)
RETURNING *;

-- name: GetFeeds :many
SELECT f.name, f.url, u.name AS username FROM feeds f JOIN users u on f.user_id = u.id;

-- name: GetFeed :one
SELECT * from feeds f WHERE f.url = $1;

-- name: MarkFeedFetched :exec
UPDATE feeds f 
SET updated_at=now(),  last_fetched_at=now()
WHERE f.id=$1;

-- name: GetNextFeedToFetch :many
SELECT *  FROM feeds f
ORDER BY f.last_fetched_at DESC NULLS FIRST;
