-- name: CreateFeedFollow :one
WITH inserted_feed_follow AS (
  INSERT INTO feed_follows (id, created_at, updated_at, user_id, feed_id)
  VALUES (gen_random_uuid(), now(), now(), $1, $2)
  RETURNING *
)
SELECT iff.id, iff.created_at, iff.updated_at, iff.user_id, iff.feed_id, u.name AS user_name, f.name AS feed_name 
FROM inserted_feed_follow iff 
JOIN users u ON iff.user_id = u.id
JOIN feeds f ON iff.feed_id = f.id;

-- name: GetFeedFollowsForUser :many
SELECT us.name AS user_name, f.name as feed_name FROM feed_follows ff 
JOIN feeds f ON ff.feed_id = f.id
JOIN users us ON  ff.user_id = us.id
WHERE ff.user_id = $1;

-- name: DeleteFeedFollow :many
DELETE FROM feed_follows 
WHERE user_id=$1 and feed_id=$2 RETURNING *;
