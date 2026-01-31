-- name: UpsertUser :one
INSERT INTO users (user_id, username, avatar_url, updated_at)
VALUES ($1, $2, $3, CURRENT_TIMESTAMP)
ON CONFLICT (user_id) DO UPDATE
SET 
    username = EXCLUDED.username,
    avatar_url = EXCLUDED.avatar_url,
    updated_at = CURRENT_TIMESTAMP
WHERE users.username IS NULL OR users.avatar_url IS NULL
RETURNING id, user_id, username, avatar_url, created_at, updated_at;

-- name: GetUserByUserID :one
SELECT id, user_id, username, avatar_url, created_at, updated_at
FROM users
WHERE user_id = $1;

-- name: GetUserByID :one
SELECT id, user_id, username, avatar_url, created_at, updated_at
FROM users
WHERE id = $1;
