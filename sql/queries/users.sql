-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, name)
VALUES (gen_random_uuid(), now(), now(), $1)
RETURNING *;

-- name: GetUser :one
SELECT * FROM users WHERE name=$1;

-- name: Resetusers :exec
DELETE FROM users;

-- name: GetUsers :many
SELECT name  from users;
